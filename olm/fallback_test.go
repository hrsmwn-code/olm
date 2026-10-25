package olm

import (
	"testing"

	"github.com/fosrl/newt/fallback"
)

func TestRelayFlagForRegistration(t *testing.T) {
	cases := []struct {
		holepunch bool
		mode      fallback.Mode
		want      bool
	}{
		{true, fallback.ModeAuto, false},
		{true, fallback.ModeUDP, false},
		{false, fallback.ModeAuto, true},
		{true, fallback.ModeTCP, true}, // tcp: always relayed, never direct UDP
	}
	for _, c := range cases {
		o := &Olm{tunnelConfig: TunnelConfig{Holepunch: c.holepunch, Fallback: fallback.Settings{Mode: c.mode}}}
		if got := o.registerRelay(o.tunnelConfig); got != c.want {
			t.Errorf("holepunch=%v mode=%s: relay=%v, want %v", c.holepunch, c.mode, got, c.want)
		}
	}
}

func TestSetupFallbackUDPModeInstallsNothing(t *testing.T) {
	fallback.Install(nil)
	o := &Olm{tunnelConfig: TunnelConfig{Endpoint: "https://p.example.com", Fallback: fallback.Settings{Mode: fallback.ModeUDP}}}
	o.setupFallback()
	defer o.closeFallback()
	if fallback.Default() != nil {
		t.Fatal("udp mode must not install a transport (upstream behaviour)")
	}
}

func TestSetupFallbackTCPModeInstalls(t *testing.T) {
	fallback.Install(nil)
	o := &Olm{tunnelConfig: TunnelConfig{Endpoint: "https://127.0.0.1:1", Fallback: fallback.Settings{Mode: fallback.ModeTCP}}}
	o.setupFallback()
	if !fallback.Default().ForceRelay() {
		t.Fatal("tcp mode must install a forcing transport")
	}
	o.closeFallback()
	if fallback.Default() != nil {
		t.Fatal("closeFallback must uninstall")
	}
}
