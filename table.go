package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Only materialize cells requested by the viewport, instead of rebuilding five
// cells for every session on each search keystroke or filter change.
type sessionContent struct {
	tview.TableContentReadOnly
	sessions []Session
	cells    map[[2]int]*tview.TableCell
}

func newSessionContent(s []Session) *sessionContent {
	return &sessionContent{sessions: s, cells: make(map[[2]int]*tview.TableCell)}
}

func (c *sessionContent) GetRowCount() int { return len(c.sessions) + 1 }

func (c *sessionContent) GetColumnCount() int { return 5 }

func (c *sessionContent) GetCell(row, col int) *tview.TableCell {
	if row < 0 || row > len(c.sessions) || col < 0 || col >= 5 {
		return nil
	}
	key := [2]int{row, col}
	if cell := c.cells[key]; cell != nil {
		return cell
	}
	var cell *tview.TableCell
	if row == 0 {
		cell = tview.NewTableCell([]string{"Status", "Updated", "Title", "Directory", "ID"}[col]).SetSelectable(false).SetTextColor(tcell.ColorYellow)
	} else {
		s := c.sessions[row-1]
		var text string
		switch col {
		case 0:
			text = state(s)
		case 1:
			text = s.Updated.Local().Format("01-02 15:04")
		case 2:
			text = titlePreview(s.Title)
		case 3:
			text = textPreview(s.CWD, 240)
		case 4:
			text = s.ID
		}
		cell = tview.NewTableCell(safe(text))
		if col == 2 {
			cell.SetExpansion(1).SetMaxWidth(48)
		}
		if col == 3 {
			cell.SetMaxWidth(35)
		}
	}
	c.cells[key] = cell
	return cell
}

// Truncate before escaping display tags, preserving complete Unicode characters.
func titlePreview(title string) string {
	return textPreview(title, 100)
}

func textPreview(title string, limit int) string {
	count := 0
	for offset := range title {
		if count == limit {
			return title[:offset]
		}
		count++
	}
	return title
}
