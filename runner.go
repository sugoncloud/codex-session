package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Runner struct{ Binary, Home string }

func (r Runner) Run(ctx context.Context, action, id string) error {
	if !uuidPattern.MatchString(id) {
		return fmt.Errorf("invalid session UUID: %q", id)
	}
	switch action {
	case "archive", "unarchive", "delete":
	default:
		return fmt.Errorf("unsupported action: %s", action)
	}
	args := []string{action, id}
	if action == "delete" {
		args = append(args, "--force")
	}
	cmd := exec.CommandContext(ctx, r.Binary, args...)
	env := []string{}
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "CODEX_HOME=") {
			env = append(env, v)
		}
	}
	cmd.Env = append(env, "CODEX_HOME="+r.Home)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("operation timed out or was canceled; refresh to verify the result: %w", ctx.Err())
		}
		return fmt.Errorf("codex %s failed: %w\n%s", action, err, strings.TrimSpace(string(output)))
	}
	return nil
}
