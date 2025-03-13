package main

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ContentBox struct {
	*tview.Box
	content tview.Primitive
}

func newContentBox(title string, content tview.Primitive) *ContentBox {
	box := tview.NewBox().SetBorder(true).SetTitle(fmt.Sprintf(" %s ", title))

	return &ContentBox{box, content}
}

func (b *ContentBox) Draw(screen tcell.Screen) {
	b.Box.DrawForSubclass(screen, b)
	x, y, w, h := b.GetInnerRect()

	b.content.SetRect(x, y, w, h)
	b.content.Draw(screen)
}

func (b *ContentBox) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return b.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		b.content.InputHandler()(event, setFocus)
	})
}
