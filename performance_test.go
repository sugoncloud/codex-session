package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func BenchmarkFilterAndRender(b *testing.B) {
	sessions := make([]Session, 10000)
	for i := range sessions {
		sessions[i] = Session{ID: fmt.Sprintf("%08d-0000-0000-0000-000000000000", i), Title: fmt.Sprintf("会话 %d 测试连接优化", i), CWD: "/opt/mdstors", Updated: time.Unix(100, 0)}
	}
	u := newUI(nil, Runner{}, "test", sessions, 0, "")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u.render("")
	}
}

func BenchmarkNavigationLongTitle(b *testing.B) {
	title := strings.Repeat("长标题包含完整提示词内容。", 10000)
	sessions := []Session{{ID: "a", Title: title, CWD: "/opt/mdstors"}, {ID: "b", Title: title, CWD: "/opt/mdstors"}}
	u := newUI(nil, Runner{}, "test", sessions, 0, "")
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		b.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(120, 32)
	u.pages.SetRect(0, 0, 120, 32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u.table.Select(i%2+1, 0)
		u.pages.Draw(screen)
	}
}
