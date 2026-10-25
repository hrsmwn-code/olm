package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fosrl/newt/fallback"
)

func TestFallbackSettingsFromFileEnvAndFlags(t *testing.T) {
	for _, k := range []string{fallback.EnvTransport, fallback.EnvRelayPath, fallback.EnvEchoPort, "CONFIG_FILE"} {
		t.Setenv(k, "")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"transport":"udp","fallbackRelayPath":"/f"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_FILE", path)

	cfg, _, _, err := LoadConfig([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Fallback.Mode != fallback.ModeUDP || cfg.Fallback.RelayPath != "/f" {
		t.Fatalf("file: %+v", cfg.Fallback)
	}

	t.Setenv(fallback.EnvTransport, "tcp")
	cfg, _, _, err = LoadConfig([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Fallback.Mode != fallback.ModeTCP {
		t.Fatalf("env: %+v", cfg.Fallback)
	}

	cfg, _, _, err = LoadConfig([]string{"--transport", "auto", "--fallback-relay-path", "/cli"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Fallback.Mode != fallback.ModeAuto || cfg.Fallback.RelayPath != "/cli" {
		t.Fatalf("cli: %+v", cfg.Fallback)
	}
}
