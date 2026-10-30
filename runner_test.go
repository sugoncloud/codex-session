package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunnerArgumentsAndHome(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "fake codex")
	output := filepath.Join(dir, "args")
	t.Setenv("TEST_RESULT", output)
	script := "#!/bin/sh\nprintf '%s\\n' \"$CODEX_HOME\" \"$@\" > \"$TEST_RESULT\"\n"
	if err := os.WriteFile(bin, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	id := "11111111-1111-4111-8111-111111111111"
	r := Runner{bin, filepath.Join(dir, "home with spaces")}
	for _, action := range []string{"archive", "unarchive", "delete"} {
		if err := r.Run(context.Background(), action, id); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		want := r.Home + "\n" + action + "\n" + id + "\n"
		if action == "delete" {
			want += "--force\n"
		}
		if string(b) != want {
			t.Fatalf("got %q want %q", b, want)
		}
	}
	if err := r.Run(context.Background(), "delete", "--all"); err == nil {
		t.Fatal("accepted invalid ID")
	}
	if err := r.Run(context.Background(), "exec", id); err == nil {
		t.Fatal("accepted unknown action")
	}
	if err := os.WriteFile(bin, []byte("#!/bin/sh\necho 'mock failure' >&2\nexit 2\n"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := r.Run(context.Background(), "archive", id); err == nil || !strings.Contains(err.Error(), "mock failure") {
		t.Fatalf("missing failure output: %v", err)
	}
}
