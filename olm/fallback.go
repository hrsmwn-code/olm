package olm

// Fork-only: TCP 443 fallback wiring for olm. Upstream files only call the
// helpers here.

import (
	"github.com/fosrl/newt/fallback"
	"github.com/fosrl/newt/logger"
	"github.com/fosrl/newt/util"
	"github.com/fosrl/olm/websocket"
)

// setupFallback installs the process-wide fallback transport for this
// tunnel. udp mode installs nothing (exact upstream behaviour). SharedBind
// (from newt) picks the transport up automatically.
func (o *Olm) setupFallback() {
	s := o.tunnelConfig.Fallback
	if err := s.Validate(); err != nil {
		logger.Error("fallback: invalid settings, using upstream UDP behaviour: %v", err)
		return
	}
	if s.Mode == fallback.ModeUDP {
		return
	}

	var opts []fallback.Option
	if o.websocket != nil {
		if tlsCfg, err := o.websocket.FallbackTLSConfig(); err != nil {
			logger.Error("fallback: TLS setup failed: %v", err)
		} else if tlsCfg != nil {
			opts = append(opts, fallback.WithTLSConfig(tlsCfg))
		}
		opts = append(opts, fallback.WithDialer(o.websocket.FallbackDialContext))
	}

	t, err := fallback.New(s, o.tunnelConfig.Endpoint, opts...)
	if err != nil {
		logger.Error("fallback: disabled: %v", err)
		return
	}
	t.OnSwitch(func(relay bool) {
		// Our source address as seen by Gerbil changes: hole punch now so
		// Pangolin updates the endpoint, and relay mode re-registers.
		if hp := o.holePunchManager; hp != nil && o.tunnelRunning {
			_ = hp.TriggerHolePunch()
			hp.ResetServerHolepunchInterval()
		}
	})
	fallback.Install(t)
}

// fallbackOnToken passes the session token and exit-node addresses to the
// transport.
func (o *Olm) fallbackOnToken(token string, exitNodeEndpoints []string) {
	t := fallback.Default()
	if t == nil {
		return
	}
	t.SetCredentials(fallback.Credentials{ClientType: "olm", ID: o.tunnelConfig.ID, Token: token})
	for _, ep := range exitNodeEndpoints {
		ips, err := util.ResolveDomainAllUpstream(ep, o.tunnelConfig.PublicDNS)
		if err != nil {
			t.AddServerHost(ep)
			continue
		}
		for _, ip := range ips {
			t.AddServerHost(ip)
		}
	}
}

// registerRelay is the "relay" flag for olm/wg/register. In tcp mode the
// client is always relayed through Gerbil: direct UDP to sites can't work.
func (o *Olm) registerRelay(cfg TunnelConfig) bool {
	return !cfg.Holepunch || cfg.Fallback.Mode == fallback.ModeTCP
}

// fallbackNetworkChanged triggers a re-probe after a socket rebind.
func (o *Olm) fallbackNetworkChanged() {
	fallback.Default().NetworkChanged()
}

func (o *Olm) closeFallback() {
	if t := fallback.Default(); t != nil {
		fallback.Install(nil)
		t.Close()
	}
}

func exitNodeEndpoints(nodes []websocket.ExitNode) []string {
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n.Endpoint)
	}
	return out
}
