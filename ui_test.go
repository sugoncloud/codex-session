package main

import (
	"context"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestUIFiltersAndDeleteCancel(t *testing.T) {
	s, err := openStore(fixture(t))
	if err != nil {
		t.Fatal(err)
	}
	defer s.db.Close()
	all, err := s.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	u := newUI(s, Runner{Binary: "must-not-execute"}, "test", all, 0, "")
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	defer screen.Fini()
	screen.SetSize(120, 32)
	u.pages.SetRect(0, 0, 120, 32)
	u.pages.Draw(screen)
	if len(u.visible) != 2 {
		t.Fatal("default must include archived")
	}
	u.search.SetText("连接")
	if len(u.visible) != 1 {
		t.Fatal("search failed")
	}
	u.search.SetText("")
	u.table.GetInputCapture()(tcell.NewEventKey(tcell.KeyTab, 0, 0))
	if len(u.visible) != 1 || u.visible[0].Archived {
		t.Fatal("active filter failed")
	}
	u.confirm("delete")
	if !u.pages.HasPage("confirm") {
		t.Fatal("missing confirmation")
	}
	u.pages.Draw(screen)
	// Delete uses a simple modal; focus defaults to Cancel, with no text entry.
	_, dialog := u.pages.GetFrontPage()
	if _, ok := dialog.(*tview.Modal); !ok {
		t.Fatal("expected a simple confirmation modal")
	}
	button, ok := u.app.GetFocus().(*tview.Button)
	if !ok || button.GetLabel() != "Cancel" {
		t.Fatal("default focus must be Cancel")
	}
	button.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, 0), func(p tview.Primitive) { u.app.SetFocus(p) })
	if u.pages.HasPage("confirm") {
		t.Fatal("cancel did not dismiss dialog")
	}
	if u.busy {
		t.Fatal("cancel executed operation")
	}
	u.search.SetText("absent")
	u.confirm("delete")
	if u.pages.HasPage("confirm") {
		t.Fatal("delete offered without selection")
	}
	// Untrusted titles must stay literal and single-line.
	if plain("a\x1b\nb") != "a  b" || safe("[red]") != tview.Escape("[red]") {
		t.Fatal("unsafe display escaping")
	}
}

func TestBusyUIStillNavigates(t *testing.T) {
	u := newUI(nil, Runner{}, "test", []Session{{ID: "one"}, {ID: "two"}}, 0, "")
	u.busy = true
	handle := u.table.GetInputCapture()
	if e := handle(tcell.NewEventKey(tcell.KeyRune, 'j', 0)); e == nil || e.Key() != tcell.KeyDown {
		t.Fatal("navigation blocked while waiting")
	}
	handle(tcell.NewEventKey(tcell.KeyRune, '/', 0))
	if u.app.GetFocus() != u.search {
		t.Fatal("search blocked while waiting")
	}
	handle(tcell.NewEventKey(tcell.KeyTab, 0, 0))
	if u.mode != 1 {
		t.Fatal("filter blocked while waiting")
	}
	u.confirm("delete")
	if u.pages.HasPage("confirm") {
		t.Fatal("concurrent mutation allowed")
	}
}

func TestTableMaterializesOnlyRequestedCells(t *testing.T) {
	c := newSessionContent(make([]Session, 10000))
	if c.GetRowCount() != 10001 || len(c.cells) != 0 {
		t.Fatal("eager cell allocation")
	}
	first := c.GetCell(1, 2)
	if c.GetCell(1, 2) != first || len(c.cells) != 1 {
		t.Fatal("cell cache not reused")
	}
	if c.GetCell(10001, 0) != nil {
		t.Fatal("out-of-bounds cell")
	}
}

func TestSelectionAfterRemoval(t *testing.T) {
	for _, tc := range []struct {
		name      string
		row       int
		remaining []Session
		want      string
	}{
		{"middle selects next", 2, []Session{{ID: "a"}, {ID: "c"}}, "c"},
		{"last selects previous", 3, []Session{{ID: "a"}, {ID: "b"}}, "b"},
		{"first selects next", 1, []Session{{ID: "b"}, {ID: "c"}}, "b"},
		{"preserves selection on reorder", 2, []Session{{ID: "c"}, {ID: "a"}, {ID: "b"}}, "b"},
		{"empty", 2, nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			u := newUI(nil, Runner{}, "test", []Session{{ID: "a"}, {ID: "b"}, {ID: "c"}}, 0, "")
			u.table.Select(tc.row, 0)
			s, _ := u.selected()
			u.all = tc.remaining
			u.render(s.ID)
			got, ok := u.selected()
			if got.ID != tc.want || ok != (tc.want != "") {
				t.Fatalf("selected %q (%v), want %q", got.ID, ok, tc.want)
			}
		})
	}
}

func TestSelectionAfterFilteredArchive(t *testing.T) {
	u := newUI(nil, Runner{}, "test", []Session{{ID: "a"}, {ID: "b"}, {ID: "c"}}, 1, "")
	u.table.Select(2, 0)
	u.all[1].Archived = true
	u.render("b")
	s, ok := u.selected()
	if !ok || s.ID != "c" {
		t.Fatalf("selected %+v", s)
	}
}
