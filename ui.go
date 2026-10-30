package main

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UI struct {
	app          *tview.Application
	pages        *tview.Pages
	table        *tview.Table
	search       *tview.InputField
	status       *tview.TextView
	store        *Store
	runner       Runner
	all, visible []Session
	mode         int
	busy         bool
}

func newUI(store *Store, runner Runner, path string, all []Session, mode int, query string) *UI {
	u := &UI{
		app:    tview.NewApplication(),
		pages:  tview.NewPages(),
		table:  tview.NewTable(),
		search: tview.NewInputField(),
		status: tview.NewTextView(),
		store:  store,
		runner: runner,
		all:    all,
		mode:   mode,
	}
	u.table.SetSelectable(true, false).SetFixed(1, 0).SetBorders(false)
	u.table.SetBorder(true)
	u.status.SetDynamicColors(false)
	u.search.SetLabel("Search / ").SetText(query).SetChangedFunc(func(_ string) { u.render("") })
	u.search.SetDoneFunc(func(k tcell.Key) {
		if k == tcell.KeyEnter || k == tcell.KeyEscape {
			u.app.SetFocus(u.table)
		}
	})
	title := tview.NewTextView().SetText("codex-sess  |  " + plain(path))
	help := tview.NewTextView().SetText("Up/Down j/k: Move  /: Search  Tab: Filter  a: Archive  u: Restore  d: Delete  r: Refresh  q: Quit")
	layout := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(title, 1, 0, false).
		AddItem(u.search, 1, 0, false).
		AddItem(u.table, 0, 1, true).
		AddItem(u.status, 2, 0, false).
		AddItem(help, 2, 0, false)
	u.pages.AddPage("main", layout, true, true)
	u.table.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Key() {
		case tcell.KeyTab:
			u.mode = (u.mode + 1) % 3
			u.render("")
			return nil
		}
		switch e.Rune() {
		case '/':
			u.app.SetFocus(u.search)
			return nil
		case 'q':
			u.app.Stop()
			return nil
		case 'r':
			u.refresh()
			return nil
		case 'a':
			u.confirm("archive")
			return nil
		case 'u':
			u.confirm("unarchive")
			return nil
		case 'd':
			u.confirm("delete")
			return nil
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, 0)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, 0)
		}
		return e
	})
	u.render("")
	u.status.SetText("Ready. All includes unarchived and archived sessions. Status does not indicate running activity.")
	return u
}

func (u *UI) Run() error {
	return u.app.SetRoot(u.pages, true).EnableMouse(true).EnablePaste(true).Run()
}

func (u *UI) selected() (Session, bool) {
	row, _ := u.table.GetSelection()
	if row < 1 || row > len(u.visible) {
		return Session{}, false
	}
	return u.visible[row-1], true
}

func (u *UI) render(keep string) {
	previousRow, _ := u.table.GetSelection()
	u.visible = filterSessions(u.all, u.search.GetText(), u.mode)
	u.table.SetContent(newSessionContent(u.visible))
	labels := []string{"All", "Unarchived", "Archived"}
	u.table.SetTitle(fmt.Sprintf(" %s | Showing %d / %d total ", labels[u.mode], len(u.visible), len(u.all)))
	selected := 1
	// A removed/filtered-out selection should stay at its old position. This
	// picks the following row, or the preceding row when the last one is gone.
	if keep != "" {
		selected = min(max(previousRow, 1), len(u.visible))
	}
	for i, s := range u.visible {
		if s.ID == keep {
			selected = i + 1
			break
		}
	}
	if len(u.visible) > 0 {
		u.table.Select(selected, 0)
	} else {
		u.table.Select(0, 0).SetOffset(0, 0)
	}
}

func (u *UI) refresh() {
	if u.busy {
		return
	}
	u.busy = true
	u.status.SetText("Refreshing...")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), readTimeout)
		defer cancel()
		all, err := u.store.List(ctx)
		u.app.QueueUpdateDraw(func() {
			u.busy = false
			if err != nil {
				u.status.SetText(plain(err.Error()))
				return
			}
			s, _ := u.selected()
			u.all = all
			u.render(s.ID)
			u.status.SetText("Refreshed")
		})
	}()
}

func (u *UI) confirm(action string) {
	if u.busy {
		return
	}
	s, ok := u.selected()
	if !ok {
		return
	}
	if action == "archive" && s.Archived {
		u.status.SetText("This session is already archived")
		return
	}
	if action == "unarchive" && !s.Archived {
		u.status.SetText("This session is not archived")
		return
	}
	name := map[string]string{"archive": "Archive", "unarchive": "Restore", "delete": "Delete permanently"}[action]
	closeDialog := func() { u.pages.RemovePage("confirm"); u.app.SetFocus(u.table) }
	message := fmt.Sprintf("%s session?\n%s\n%s", name, safe(s.Title), s.ID)
	if action == "delete" {
		message = fmt.Sprintf("UUID: %s\nTitle: %s", s.ID, safe(titlePreview(s.Title)))
	}
	modal := tview.NewModal().SetText(message).AddButtons([]string{"Cancel", name}).SetDoneFunc(func(i int, _ string) {
		closeDialog()
		if i == 1 {
			u.perform(action, s)
		}
	})
	u.pages.AddPage("confirm", modal, true, true)
	u.app.SetFocus(modal)
}

func (u *UI) perform(action string, s Session) {
	if u.busy {
		return
	}
	u.busy = true
	u.status.SetText("Running " + action + " ...")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
		err := u.runner.Run(ctx, action, s.ID)
		cancel()
		readCtx, readCancel := context.WithTimeout(context.Background(), readTimeout)
		all, readErr := u.store.List(readCtx)
		readCancel()
		u.app.QueueUpdateDraw(func() {
			u.busy = false
			if readErr == nil {
				selected, _ := u.selected()
				u.all = all
				u.render(selected.ID)
			}
			switch {
			case err != nil:
				u.status.SetText(plain(err.Error()))
			case readErr != nil:
				u.status.SetText("Operation completed, but refresh failed: " + plain(readErr.Error()))
			default:
				u.status.SetText("Completed: " + action + " " + s.ID)
			}
		})
	}()
}
