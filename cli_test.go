package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIListAndJSON(t *testing.T) {
	home := filepath.Dir(fixture(t))
	var out bytes.Buffer
	if err := runCLI([]string{"--home", home, "--json", "--status", "archived"}, &out); err != nil {
		t.Fatal(err)
	}
	var sessions []Session
	if err := json.Unmarshal(out.Bytes(), &sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 1 || !sessions[0].Archived {
		t.Fatalf("unexpected output: %s", out.String())
	}
	out.Reset()
	if err := runCLI([]string{"--home", home, "--list", "--search", "连接"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Unarchived") || strings.Contains(out.String(), "归档测试") {
		t.Fatalf("unexpected table: %s", out.String())
	}
}

func TestCLIHelpVersionAndInvalidOptions(t *testing.T) {
	for _, tc := range []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{"version", []string{"--version"}, "codex-sess", false},
		{"invalid status", []string{"--status", "unknown"}, "", true},
		{"unexpected positional", []string{"unexpected"}, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			err := runCLI(tc.args, &out)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v", err)
			}
			if !strings.Contains(out.String(), tc.want) {
				t.Fatalf("output = %s", out.String())
			}
		})
	}
	var out bytes.Buffer
	if err := runCLI([]string{"--help"}, &out); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("help error = %v", err)
	}
	if !strings.Contains(out.String(), "Codex data directory") {
		t.Fatal("help output missing")
	}
}
