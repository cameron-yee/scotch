package main

import (
        "fmt"
        "strings"
        "strconv"

        "github.com/rivo/tview"
        "github.com/gdamore/tcell/v2"
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
        	if (
                	(textFilter == "" ||
                	strings.Contains(strings.ToLower(entry), strings.ToLower(textFilter))) &&
                	int(item.Price) >= uiState.PriceMinimum &&
                	int(item.Price) <= uiState.PriceMaximum) {
                	builder.WriteString(entry)
        	}
        }
        text := builder.String()
        return text
}

func configureTextView(data *Data, uiState *UIState) *tview.TextView {
	textview := tview.NewTextView()
	initialText := getVisibleItems(data, uiState)

	textview.SetText(initialText)
	textview.ScrollToBeginning()

	return textview
}

func configureTextFilter(data *Data, uiState *UIState, textView *tview.TextView) *tview.InputField {
	textFilter := tview.NewInputField().
		SetLabel("Text Filter: ").
		SetFieldWidth(10)

	textFilter.SetChangedFunc(func(text string) {
        	uiState.Text = text
        	textView.SetText(getVisibleItems(data, uiState))
	})

	return textFilter
}

func configurePriceMinimumFilter(data *Data, uiState *UIState, textView *tview.TextView) *tview.InputField {
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
        	textView.SetText(getVisibleItems(data, uiState))
	})

	return priceMinimumFilter
}

func configurePriceMaximumFilter(data *Data, uiState *UIState, textView *tview.TextView) *tview.InputField {
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
        	textView.SetText(getVisibleItems(data, uiState))
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

func render(data *Data) {
	uiState := &UIState{
        	PriceMaximum: 99999,
        	PriceMinimum: 0,
        	Text: "",
        	FocusedIndex: 0,
	}
	app := tview.NewApplication()

	textView := configureTextView(data, uiState)
	textFilter := configureTextFilter(data, uiState, textView)
	priceMinimumFilter := configurePriceMinimumFilter(data, uiState, textView)
	priceMaximumFilter := configurePriceMaximumFilter(data, uiState, textView)

	fields := [4]tview.Primitive{textFilter, priceMinimumFilter, priceMaximumFilter, textView}
	handleNavigation(app, uiState, fields)

	flex := tview.NewFlex().
 		AddItem(textFilter, 0, 1, true).
 		AddItem(priceMinimumFilter, 0, 1, false).
 		AddItem(priceMaximumFilter, 0, 1, false).
 		AddItem(textView, 0, 12, false).
 		SetDirection(0)

        if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
        	panic(err)
        }
}
