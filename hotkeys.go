package main

import (
	"fmt"
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// HotKey represents a keyboard shortcut with description
type HotKey struct {
	desc     string
	shortcut string
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
func (r *HotKeys) AddHotKey(desc string, shortcut string) *HotKeys {
	r.values = append(r.values, HotKey{desc, shortcut})

	return r
}

// Draw renders the hotkeys component
func (r *HotKeys) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()

	// Calculate how many columns we need
	totalHotkeys := len(r.values)
	if totalHotkeys == 0 {
		return
	}

	// Determine max hotkey text length for column width calculation
	maxHotkeyLength := 0
	for _, hotkey := range r.values {
		hotkeyText := fmt.Sprintf("<%s> %s", string(hotkey.shortcut), hotkey.desc)
		if len(hotkeyText) > maxHotkeyLength {
			maxHotkeyLength = len(hotkeyText)
		}
	}

	// Add some padding between columns
	columnWidth := maxHotkeyLength + 4

	// Calculate how many columns can fit in the available width
	maxColumns := int(math.Max(1, float64(width)/float64(columnWidth)))

	// Calculate how many rows we need per column
	rowsPerColumn := int(math.Ceil(float64(totalHotkeys) / float64(maxColumns)))

	// Ensure we don't exceed available height
	if rowsPerColumn > height {
		rowsPerColumn = height
		maxColumns = int(math.Ceil(float64(totalHotkeys) / float64(rowsPerColumn)))
	}

	// Draw hotkeys in columns
	for i, hotkey := range r.values {
		// Calculate column and row position
		column := i / rowsPerColumn
		row := i % rowsPerColumn

		// Skip if we've run out of columns that can fit in the width
		if column >= maxColumns {
			break
		}

		// Calculate x position for this column
		colX := x + (column * columnWidth)

		// Draw the hotkey
		line := fmt.Sprintf("<%s> %s", string(hotkey.shortcut), hotkey.desc)
		tview.Print(screen, line, colX, y+row, columnWidth, tview.AlignLeft, tcell.ColorYellow)
	}
}

// GetConnectionHotkeys returns hotkeys for the connections view
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

// GetDatabaseHotkeys returns hotkeys for the database view
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
		AddHotKey("Help", "h").
		AddHotKey("Quit", "q").
		AddHotKey("Up", "↑").
		AddHotKey("Down", "↓").
		AddHotKey("Left", "←").
		AddHotKey("Right", "→")
}

// GetHelpHotkeys returns hotkeys for the help view
func GetHelpHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("Back", "b").
		AddHotKey("Scroll Up", "↑").
		AddHotKey("Scroll Down", "↓").
		AddHotKey("Quit", "q")
}

// GetFormHotkeys returns hotkeys for form views
func GetFormHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("Next Field", "Tab").
		AddHotKey("Previous Field", "Shift+Tab").
		AddHotKey("Submit", "Enter").
		AddHotKey("Cancel", "Esc")
}

// GetQueryHotkeys returns hotkeys for the query editor view
func GetQueryHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("Execute", "Ctrl+e").
		AddHotKey("Save Query", "Ctrl+s").
		AddHotKey("Load Query", "Ctrl+o").
		AddHotKey("Clear", "Ctrl+l").
		AddHotKey("Exit Editor", "Esc").
		AddHotKey("History", "Ctrl+h")
}

// GetExportHotkeys returns hotkeys for the export view
func GetExportHotkeys() *HotKeys {
	return NewHotkeys().
		AddHotKey("Export as CSV", "c").
		AddHotKey("Export as JSON", "j").
		AddHotKey("Export as SQL", "s").
		AddHotKey("Cancel", "Esc")
}
