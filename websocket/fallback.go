package websocket

import (
	"context"
	"crypto/tls"
	"net"
	"os"
)

// Fork-only (TCP 443 fallback) accessors so the fallback relay connection
// uses the same TLS settings and dialer (physical-network DNS, gateway
// bypass reporting) as the control WebSocket.

// FallbackTLSConfig returns the TLS config for the relay, or nil for default.
func (c *Client) FallbackTLSConfig() (*tls.Config, error) {
	var cfg *tls.Config
	if c.tlsConfig.ClientCertFile != "" || c.tlsConfig.ClientKeyFile != "" || len(c.tlsConfig.CAFiles) > 0 || c.tlsConfig.PKCS12File != "" {
		var err error
		if cfg, err = c.setupTLS(); err != nil {
			return nil, err
		}
	}
	if os.Getenv("SKIP_TLS_VERIFY") == "true" {
		if cfg == nil {
			cfg = &tls.Config{}
		}
		cfg.InsecureSkipVerify = true
	}
	return cfg, nil
}

// FallbackDialContext dials like the control WebSocket: resolve via the
// physical network's DNS and report the target IPs via OnDialTargets, so
// full-tunnel mode keeps them on bypass routes.
func (c *Client) FallbackDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return c.dialContext(ctx, network, addr)
}
