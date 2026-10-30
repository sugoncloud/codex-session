package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// version can be overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := runCLI(os.Args[1:], os.Stdout); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, "codex-sess:", err)
		os.Exit(1)
	}
}
