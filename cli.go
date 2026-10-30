package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/tabwriter"
)

func runCLI(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("codex-sess", flag.ContinueOnError)
	flags.SetOutput(output)
	defaultHome := os.Getenv("CODEX_HOME")
	if defaultHome == "" {
		h, e := os.UserHomeDir()
		if e != nil {
			return e
		}
		defaultHome = filepath.Join(h, ".codex")
	}
	home := flags.String("home", defaultHome, "Codex data directory (also used for session operations)")
	binary := flags.String("codex", "codex", "Path to the Codex CLI")
	list := flags.Bool("list", false, "Print sessions and exit")
	jsonOut := flags.Bool("json", false, "Print sessions as JSON and exit")
	status := flags.String("status", "all", "all / active (unarchived) / archived")
	search := flags.String("search", "", "Search titles, IDs, and directories")
	ver := flags.Bool("version", false, "Print version")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *ver {
		fmt.Fprintln(output, "codex-sess", version)
		return nil
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("positional arguments are not supported; use --help")
	}
	mode := 0
	switch *status {
	case "all":
	case "active":
		mode = 1
	case "archived":
		mode = 2
	default:
		return fmt.Errorf("invalid --status: %s", *status)
	}
	abs, err := filepath.Abs(*home)
	if err != nil {
		return err
	}
	path, err := discoverDB(abs)
	if err != nil {
		return err
	}
	store, err := openStore(path)
	if err != nil {
		return err
	}
	defer store.Close()
	ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
	sessions, err := store.List(ctx)
	cancel()
	if err != nil {
		return err
	}
	if *list || *jsonOut {
		filtered := filterSessions(sessions, *search, mode)
		if *jsonOut {
			enc := json.NewEncoder(output)
			enc.SetIndent("", "  ")
			return enc.Encode(filtered)
		}
		w := tabwriter.NewWriter(output, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tStatus\tUpdated\tTitle\tDirectory")
		for _, s := range filtered {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.ID, state(s), s.Updated.Local().Format("2006-01-02 15:04"), plain(s.Title), plain(s.CWD))
		}
		return w.Flush()
	}
	ui := newUI(store, Runner{Binary: *binary, Home: abs}, path, sessions, mode, *search)
	return ui.Run()
}
