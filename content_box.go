package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ContentBox struct {
	*tview.Flex
	content   tview.Primitive
	box       *tview.Box
	searchBar *tview.InputField
}

func newContentBox(title string, content tview.Primitive) *ContentBox {
	box := tview.NewBox().SetBorder(true).SetTitle(fmt.Sprintf(" %s ", title))

	container := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(box, 0, 1, true)

	return &ContentBox{container, content, box, nil}
}

func (b *ContentBox) toggleSearchBar(searchFunc func(value string)) {
	if b.searchBar != nil {
		searchBar := b.GetItem(0)
		b.searchBar = nil
		b.RemoveItem(searchBar)
		return
	}

	// Add the search bar re-adding the content so ordering is preserved
	searchBar := newSearchBar(searchFunc)
	b.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		searchBar.InputHandler()(event, func(p tview.Primitive) {})
		return event
	})
	b.searchBar = searchBar
	content := b.GetItem(0)
	b.RemoveItem(content)
	b.AddItem(searchBar, 3, 0, true)
	b.AddItem(content, 0, 1, false)
}

func (b *ContentBox) Draw(screen tcell.Screen) {
	b.Flex.Draw(screen)

	x, y, w, h := b.box.GetInnerRect()

	b.content.SetRect(x, y, w, h)
	b.content.Draw(screen)
}

func (b *ContentBox) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if b.searchBar != nil {
			return
		}

		b.content.InputHandler()(event, setFocus)
	})
}

func newSearchBar(apply func(value string)) *tview.InputField {
	searchBar := tview.NewInputField().
		SetFieldWidth(0).
		SetChangedFunc(apply)
	searchBar.SetBorder(true)
	searchBar.SetFieldBackgroundColor(tcell.ColorNone)

	return searchBar
}
