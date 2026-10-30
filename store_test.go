package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func fixture(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "state_5.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE threads (id TEXT PRIMARY KEY,title TEXT,cwd TEXT,archived INTEGER,updated_at INTEGER);
 INSERT INTO threads VALUES ('11111111-1111-4111-8111-111111111111','连接优化','/opt/mdstors',0,100),('22222222-2222-4222-8222-222222222222','归档测试','/opt/local',1,200);`)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func TestStoreReadOnlyAndFilters(t *testing.T) {
	path := fixture(t)
	s, err := openStore(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s.db.Close()
	list, err := s.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || !list[0].Archived {
		t.Fatalf("unexpected sessions: %+v", list)
	}
	for _, tc := range []struct {
		query      string
		mode, want int
	}{{"", 0, 2}, {"", 1, 1}, {"", 2, 1}, {"MDSTORS", 0, 1}, {"归档", 2, 1}, {"absent", 0, 0}} {
		if n := len(filterSessions(list, tc.query, tc.mode)); n != tc.want {
			t.Errorf("filter %+v: got %d", tc, n)
		}
	}
	if _, err = s.db.Exec("DELETE FROM threads"); err == nil {
		t.Fatal("database allowed mutation")
	}
	list, err = s.List(context.Background())
	if err != nil || len(list) != 2 {
		t.Fatalf("read-only check damaged data: %v", err)
	}
}

func TestDiscoverNumericVersionAndMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := discoverDB(dir); err == nil {
		t.Fatal("missing database accepted")
	}
	for _, n := range []string{"state_5.sqlite", "state_12.sqlite", "state_9.sqlite"} {
		if err := os.WriteFile(filepath.Join(dir, n), nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	p, err := discoverDB(dir)
	if err != nil || filepath.Base(p) != "state_12.sqlite" {
		t.Fatalf("%s %v", p, err)
	}
}

func TestOpenDoesNotCreateDatabase(t *testing.T) {
	p := filepath.Join(t.TempDir(), "missing.sqlite")
	if _, err := openStore(p); err == nil {
		t.Fatal("expected missing error")
	}
	if _, err := os.Stat(p); !os.IsNotExist(err) {
		t.Fatal("created missing DB")
	}
}
