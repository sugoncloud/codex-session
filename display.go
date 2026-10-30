package main

import (
	"strings"
	"unicode"

	"github.com/rivo/tview"
)

func plain(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
}

func safe(s string) string { return tview.Escape(plain(s)) }

func state(s Session) string {
	if s.Archived {
		return "Archived"
	}
	return "Unarchived"
}
