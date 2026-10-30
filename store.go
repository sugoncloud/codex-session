package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct{ db *sql.DB }

func (s *Store) Close() error {
	return s.db.Close()
}

func discoverDB(home string) (string, error) {
	files, err := filepath.Glob(filepath.Join(home, "state_*.sqlite"))
	if err != nil {
		return "", err
	}
	best, n := "", -1
	for _, f := range files {
		base := filepath.Base(f)
		v, e := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(base, "state_"), ".sqlite"))
		if e == nil && v > n {
			best, n = f, v
		}
	}
	if best == "" {
		return "", fmt.Errorf("no state_*.sqlite found in %s; check --home", home)
	}
	return best, nil
}

func openStore(path string) (*Store, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	st, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !st.Mode().IsRegular() {
		return nil, fmt.Errorf("not a database file: %s", abs)
	}
	u := url.URL{Scheme: "file", Path: abs}
	q := u.Query()
	q.Set("mode", "ro")
	q.Add("_pragma", "busy_timeout(3000)")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	return &Store{db: db}, nil
}

func (s *Store) List(ctx context.Context) ([]Session, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, COALESCE(title,''), COALESCE(cwd,''), archived, updated_at FROM threads ORDER BY updated_at DESC, id")
	if err != nil {
		return nil, fmt.Errorf("cannot read sessions (database schema may be incompatible): %w", err)
	}
	defer rows.Close()
	sessions := []Session{}
	for rows.Next() {
		var x Session
		var archived int
		var updated int64
		if err := rows.Scan(&x.ID, &x.Title, &x.CWD, &archived, &updated); err != nil {
			return nil, err
		}
		x.Archived = archived != 0
		x.Updated = time.Unix(updated, 0)
		sessions = append(sessions, x)
	}
	return sessions, rows.Err()
}
