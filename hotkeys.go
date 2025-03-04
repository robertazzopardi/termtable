package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HotKey represents a keyboard shortcut with description
type HotKey struct {
	desc     string
	shortcut rune
}

// HotKeys is a UI component for displaying keyboard shortcuts
type HotKeys struct {
	*tview.List
	values []HotKey
}

// NewHotkeys creates a new HotKeys component
func NewHotkeys() *HotKeys {
	list := tview.NewList().
		ShowSecondaryText(false).SetSelectedFocusOnly(true)

	return &HotKeys{
		List:   list,
		values: []HotKey{},
	}
}

// AddHotKey adds a new hotkey to the component
func (r *HotKeys) AddHotKey(desc string, shortcut rune) *HotKeys {
	r.values = append(r.values, HotKey{desc, shortcut})

	return r
}

// Draw renders the hotkeys component
func (r *HotKeys) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()

	for index, hotkey := range r.values {
		if index >= height {
			break
		}

		line := fmt.Sprintf("<%s> %s", string(hotkey.shortcut), hotkey.desc)
		tview.Print(screen, line, x, y+index, width, tview.AlignLeft, tcell.ColorYellow)
	}
}
