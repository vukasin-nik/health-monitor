package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "targets.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestLoadConfig_AppliesDefaults(t *testing.T) {
	path := writeConfig(t, `{"targets":[{"name":"a","url":"http://example.com"}]}`)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IntervalSeconds != 10 || cfg.TimeoutSeconds != 5 {
		t.Errorf("defaults not applied: %+v", cfg)
	}
}

func TestLoadConfig_Errors(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"no targets", `{"targets":[]}`},
		{"missing name", `{"targets":[{"url":"http://example.com"}]}`},
		{"invalid url", `{"targets":[{"name":"a","url":"not-a-url"}]}`},
		{"bad json", `{`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := LoadConfig(writeConfig(t, tc.content)); err == nil {
				t.Error("expected an error, got nil")
			}
		})
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	if _, err := LoadConfig("does-not-exist.json"); err == nil {
		t.Error("expected an error for missing file")
	}
}
