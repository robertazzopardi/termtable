package main

import "github.com/rivo/tview"

type Layout struct {
	*tview.Flex
	header  *tview.Flex
	content *tview.Pages
}

func newLayout(direction int, header *tview.Flex, content *tview.Pages) Layout {
	view := tview.NewFlex().SetDirection(direction).
		AddItem(header, 0, 1, false).
		AddItem(content, 0, 6, false)

	return Layout{view, header, content}
}
