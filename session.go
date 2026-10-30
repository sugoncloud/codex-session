package main

import (
	"strings"
	"time"
)

const (
	readTimeout      = 5 * time.Second
	operationTimeout = 30 * time.Second
)

type Session struct {
	ID       string    `json:"id"`
	Title    string    `json:"title"`
	CWD      string    `json:"cwd"`
	Archived bool      `json:"archived"`
	Updated  time.Time `json:"updated"`
}

func filterSessions(all []Session, query string, mode int) []Session {
	out := []Session{}
	query = strings.ToLower(strings.TrimSpace(query))
	for _, s := range all {
		if mode == 1 && s.Archived || mode == 2 && !s.Archived {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(s.Title+"\n"+s.ID+"\n"+s.CWD), query) {
			continue
		}
		out = append(out, s)
	}
	return out
}
