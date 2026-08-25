package config

import "testing"

func TestConfigExportDirDefault(t *testing.T) {
	t.Setenv("EXPORT_DIR", "")
	cfg := Load()
	if cfg.ExportDir == "" {
		t.Fatal("empty export dir default")
	}
}
