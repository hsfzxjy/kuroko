package tui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type SelectorData interface {
	GetTexts() (string, string)
	SetRenderFunc(func())
}

type itemWrapper struct {
	app   *SelectorApp
	Data  SelectorData
	Index int
}

func (iw *itemWrapper) Render() {
	t, st := iw.Data.GetTexts()
	iw.app.QueueUpdateDraw(func() {
		iw.app.List.SetItemText(iw.Index, t, st)
	})
}

type SelectorApp struct {
	*tview.Application
	List      *tview.List
	StatusBar *tview.TextView

	Items []*itemWrapper
}

func NewSelectorApp() *SelectorApp {
	app := new(SelectorApp)
	app.List = tview.NewList().
		SetSecondaryTextColor(tcell.ColorWhiteSmoke).
		SetHighlightFullLine(true)
	app.List.SetBorder(true).
		SetBorderAttributes(tcell.AttrDim)
	app.StatusBar = tview.NewTextView()

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(app.List, 0, 1, true).
		AddItem(app.StatusBar, 1, 0, false)
	app.Application = tview.NewApplication().
		SetRoot(flex, true).
		SetFocus(flex)

	return app
}

func (sa *SelectorApp) AddItem(data SelectorData, onSelect func()) {
	item := &itemWrapper{
		app:   sa,
		Data:  data,
		Index: sa.List.GetItemCount(),
	}
	data.SetRenderFunc(item.Render)
	sa.Items = append(sa.Items, item)
	sa.List.AddItem("", "", '\u0000', onSelect)
	item.Render()
}
