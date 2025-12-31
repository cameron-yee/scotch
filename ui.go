package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type UIState struct {
	Text         string
	PriceMaximum int
	PriceMinimum int
	FocusedIndex int
}

func getVisibleItems(root *Data, uiState *UIState) string {
	textFilter := uiState.Text

	var builder strings.Builder
	for _, item := range root.Items {
		entry := fmt.Sprintf("%s -- $%.2f    (UPC: %d)\n", item.Item, item.Price, item.UPC)
		if (textFilter == "" ||
			strings.Contains(strings.ToLower(entry), strings.ToLower(textFilter))) &&
			int(item.Price) >= uiState.PriceMinimum &&
			int(item.Price) <= uiState.PriceMaximum {
			builder.WriteString(entry)
		}
	}
	text := builder.String()
	return text
}

func getNewItems(AppData *AppData, uiState *UIState) string {
	textFilter := uiState.Text

	var builder strings.Builder
	for _, a := range AppData.latestData.data.Items {
		isNew := true
		for _, b := range AppData.previousData.data.Items {
			if a.Item == b.Item {
				isNew = false
				break
			}
		}

		if !isNew {
			continue
		}

		entry := fmt.Sprintf("%s -- $%.2f    (UPC: %d)\n", a.Item, a.Price, a.UPC)
		if (textFilter == "" ||
			strings.Contains(strings.ToLower(entry), strings.ToLower(textFilter))) &&
			int(a.Price) >= uiState.PriceMinimum &&
			int(a.Price) <= uiState.PriceMaximum {
			builder.WriteString(entry)
		}
	}
	text := builder.String()
	return text
}

func configurePreviousView(lastUpdated *time.Time, uiState *UIState) *tview.TextView {
	textview := tview.NewTextView()
	formattedDate := lastUpdated.Format("01/02/06")

	textview.SetText(fmt.Sprintf("New Since: %s", formattedDate))

	return textview
}

func configureNewItemsView(appData *AppData, uiState *UIState) *tview.TextView {
	textview := tview.NewTextView()
	newItems := getNewItems(appData, uiState)

	textview.SetText(newItems)

	return textview
}

func configureLastUpdatedView(lastUpdated *time.Time, uiState *UIState) *tview.TextView {
	textview := tview.NewTextView()
	formattedDate := lastUpdated.Format("01/02/06")

	textview.SetText(fmt.Sprintf("Last Updated: %s", formattedDate))

	return textview
}

func configureDataView(data *Data, uiState *UIState) *tview.TextView {
	textview := tview.NewTextView()
	initialText := getVisibleItems(data, uiState)

	textview.SetText(initialText)
	textview.ScrollToBeginning()

	return textview
}

func configureTextFilter(appData *AppData, uiState *UIState, textView *tview.TextView, newItemsView *tview.TextView) *tview.InputField {
	textFilter := tview.NewInputField().
		SetLabel("Text Filter: ").
		SetFieldWidth(10)

	textFilter.SetChangedFunc(func(text string) {
		uiState.Text = text
		textView.SetText(getVisibleItems(appData.latestData.data, uiState))
		newItemsView.SetText(getNewItems(appData, uiState))
	})

	return textFilter
}

func configurePriceMinimumFilter(appData *AppData, uiState *UIState, textView *tview.TextView, newItemsView *tview.TextView) *tview.InputField {
	priceMinimumFilter := tview.NewInputField().
		SetLabel("Minimum Price: ").
		SetFieldWidth(5).
		SetAcceptanceFunc(tview.InputFieldInteger)

	priceMinimumFilter.SetChangedFunc(func(text string) {
		priceMinimum, err := strconv.Atoi(text)
		if err != nil {
			return
		}
		uiState.PriceMinimum = priceMinimum
		textView.SetText(getVisibleItems(appData.latestData.data, uiState))
		newItemsView.SetText(getNewItems(appData, uiState))
	})

	return priceMinimumFilter
}

func configurePriceMaximumFilter(appData *AppData, uiState *UIState, textView *tview.TextView, newItemsView *tview.TextView) *tview.InputField {
	priceMaximumFilter := tview.NewInputField().
		SetLabel("Maximum Price: ").
		SetFieldWidth(5).
		SetAcceptanceFunc(tview.InputFieldInteger)

	priceMaximumFilter.SetChangedFunc(func(text string) {
		priceMaximum, err := strconv.Atoi(text)
		if err != nil {
			return
		}
		uiState.PriceMaximum = priceMaximum
		textView.SetText(getVisibleItems(appData.latestData.data, uiState))
		newItemsView.SetText(getNewItems(appData, uiState))
	})

	return priceMaximumFilter
}

func handleNavigation(app *tview.Application, uiState *UIState, fields [4]tview.Primitive) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB, tcell.KeyCtrlN:
			uiState.FocusedIndex += 1
			if uiState.FocusedIndex == 4 {
				uiState.FocusedIndex = 0
			}
			app.SetFocus(fields[uiState.FocusedIndex])
		case tcell.KeyCtrlP:
			uiState.FocusedIndex -= 1
			if uiState.FocusedIndex == -1 {
				uiState.FocusedIndex = 3
			}
			app.SetFocus(fields[uiState.FocusedIndex])
		}

		return event
	})
}

func render(appData *AppData) {
	uiState := &UIState{
		PriceMaximum: 99999,
		PriceMinimum: 0,
		Text:         "",
		FocusedIndex: 0,
	}
	app := tview.NewApplication()

	lastUpdatedView := configureLastUpdatedView(appData.latestData.updatedDate, uiState)
	textView := configureDataView(appData.latestData.data, uiState)
	previousView := configurePreviousView(appData.previousData.updatedDate, uiState)
	newItemsView := configureNewItemsView(appData, uiState)
	textFilter := configureTextFilter(appData, uiState, textView, newItemsView)
	priceMinimumFilter := configurePriceMinimumFilter(appData, uiState, textView, newItemsView)
	priceMaximumFilter := configurePriceMaximumFilter(appData, uiState, textView, newItemsView)

	fields := [4]tview.Primitive{textFilter, priceMinimumFilter, priceMaximumFilter, textView}
	handleNavigation(app, uiState, fields)

	leftColumn := tview.NewFlex().
		AddItem(lastUpdatedView, 0, 1, false).
		AddItem(textFilter, 0, 1, true).
		AddItem(priceMinimumFilter, 0, 1, false).
		AddItem(priceMaximumFilter, 0, 1, false).
		AddItem(textView, 0, 12, false).
		SetDirection(0)

	rightColumn := tview.NewFlex().
		AddItem(previousView, 0, 1, false).
		AddItem(newItemsView, 0, 12, false).
		SetDirection(0)

	flex := tview.NewFlex().
		AddItem(leftColumn, 0, 1, true).
		AddItem(rightColumn, 0, 1, false).
		SetDirection(1)

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
