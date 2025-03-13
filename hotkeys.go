package main

import (
	"fmt"
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type HotKey struct {
	desc     string
	shortcut string
}

type HotKeys struct {
	*tview.List
	values []HotKey
}

func NewHotkeys() *HotKeys {
	list := tview.NewList().
		ShowSecondaryText(false).SetSelectedFocusOnly(true)

	return &HotKeys{
		List:   list,
		values: []HotKey{},
	}
}

func (r *HotKeys) AddHotKey(desc string, shortcut string) *HotKeys {
	r.values = append(r.values, HotKey{desc, shortcut})

	return r
}

func (r *HotKeys) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()

	totalHotkeys := len(r.values)
	if totalHotkeys == 0 {
		return
	}

	maxHotkeyLength := 0
	for _, hotkey := range r.values {
		hotkeyText := fmt.Sprintf("<%s> %s", string(hotkey.shortcut), hotkey.desc)
		if len(hotkeyText) > maxHotkeyLength {
			maxHotkeyLength = len(hotkeyText)
		}
	}

	columnWidth := maxHotkeyLength + 4

	maxColumns := int(math.Max(1, float64(width)/float64(columnWidth)))

	rowsPerColumn := int(math.Ceil(float64(totalHotkeys) / float64(maxColumns)))

	if rowsPerColumn > height {
		rowsPerColumn = height
		maxColumns = int(math.Ceil(float64(totalHotkeys) / float64(rowsPerColumn)))
	}

	for i, hotkey := range r.values {
		column := i / rowsPerColumn
		row := i % rowsPerColumn

		if column >= maxColumns {
			break
		}

		colX := x + (column * columnWidth)

		line := fmt.Sprintf("<%s> %s", string(hotkey.shortcut), hotkey.desc)
		tview.Print(screen, line, colX, y+row, columnWidth, tview.AlignLeft, tcell.ColorYellow)
	}
}

func GetConnectionHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("New Connection", "n").
		AddHotKey("Edit Connection", "e").
		AddHotKey("Delete Connection", "d").
		AddHotKey("Open Connection", "o").
		AddHotKey("Test Connection", "t").
		AddHotKey("Refresh Connections", "r").
		AddHotKey("Search", "/").
		AddHotKey("Sort by Name", "s").
		AddHotKey("Help", "h").
		AddHotKey("Quit", "q").
		AddHotKey("Up", "↑").
		AddHotKey("Down", "↓").
		AddHotKey("Enter", "⏎")
}

func GetDatabaseHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("Back to Connections", "b").
		AddHotKey("Refresh Data", "r").
		AddHotKey("Execute Query", "e").
		AddHotKey("Export Results", "x").
		AddHotKey("Filter Results", "f").
		AddHotKey("Copy Row", "c").
		AddHotKey("Copy Cell", "y").
		AddHotKey("Next Page", "n").
		AddHotKey("Previous Page", "p").
		AddHotKey("Toggle View Mode", "v").
		AddHotKey("Help", "?").
		AddHotKey("Quit", "q").
		AddHotKey("Up", "↑").
		AddHotKey("Down", "↓").
		AddHotKey("Left", "←").
		AddHotKey("Right", "→")
}

func GetHelpHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("Back", "b").
		AddHotKey("Scroll Up", "↑").
		AddHotKey("Scroll Down", "↓").
		AddHotKey("Quit", "q")
}

func GetFormHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("Next Field", "Tab").
		AddHotKey("Previous Field", "Shift+Tab").
		AddHotKey("Submit", "Enter").
		AddHotKey("Cancel", "Esc")
}

func GetQueryHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("Execute", "Ctrl+e").
		AddHotKey("Save Query", "Ctrl+s").
		AddHotKey("Load Query", "Ctrl+o").
		AddHotKey("Clear", "Ctrl+l").
		AddHotKey("Exit Editor", "Esc").
		AddHotKey("History", "Ctrl+h")
}

func GetExportHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("Export as CSV", "c").
		AddHotKey("Export as JSON", "j").
		AddHotKey("Export as SQL", "s").
		AddHotKey("Cancel", "Esc")
}
