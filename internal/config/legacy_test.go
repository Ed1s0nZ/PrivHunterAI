package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLegacyImport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	os.WriteFile(path, []byte(`{"AI":"gpt","apiKeys":{"gpt":"fixture"},"headers2":{"Cookie":"fixture"}}`), 0600)
	v, e := ImportLegacy(path)
	if e != nil || !v.Paused || v.Endpoint != "https://api.openai.com/v1/chat/completions" || v.APIKey != "fixture" {
		t.Fatal("migration failed", e)
	}
}
