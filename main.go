// Demo code for the List primitive.
package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/rivo/tview"
)

type Root struct {
	XMLName xml.Name  `xml:"Root"` // Maps to the <catalog> element
	Items []Inventory `xml:"Inventory"`    // Maps to a slice of <book> elements
}
type Inventory struct {
	XMLName  xml.Name `xml:"Inventory"` // Maps to the <Inventory> element
	Category string   `xml:"Category"` // Maps to the 'id' attribute of <book>
	Item     string   `xml:"Item"`
	Price    float64  `xml:"Price"`
	UPC      int64    `xml:"UPC"`
}

func getDateURL(daysAgo int) string {
        now := time.Now()
        targetDate := now.AddDate(0, 0, -daysAgo)
        formattedTargetDate := targetDate.Format("01_02_06")

	// https://www.dtnfspiritsandwines.com/AllLiquor_11_04_25.xml
        return fmt.Sprintf("https://www.dtnfspiritsandwines.com/AllLiquor_%s.xml", formattedTargetDate)
}

func fetch(url string) (string, error) {
        resp, err := http.Get(url)
        if err != nil {
                return "", err
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
        log.Fatal(err)
        }
        xmlData := string(body)
        return xmlData, nil
}

//
// Tries to get the correct URL with the latest date that has the XML data.
// Will try 30 times before giving up.
//
func getXMLData() *Root {
        tries := 0

	for tries < 31 {
                url := getDateURL(tries)
                xmlData, err := fetch(url)
                if err != nil {
                        tries += 1
                        continue
                }

                var root Root
        	err2 := xml.Unmarshal([]byte(xmlData), &root)
        	if err2 != nil {
                	tries += 1
                	continue
        	}

                return &root
	}

	panic("Could not find valid URL in 30 tries")
}

func getVisibleItems(root *Root, filter string) string {
        var builder strings.Builder
        for _, item := range root.Items {
        	entry := fmt.Sprintf("%s -- $%.2f    (UPC: %d)\n", item.Item, item.Price, item.UPC)
        	if (filter == "" || strings.Contains(strings.ToLower(entry), strings.ToLower(filter))) {
                	builder.WriteString(entry)
        	}
        }
        text := builder.String()
        return text
}

func main() {
        root := getXMLData()

	app := tview.NewApplication()
	textview := tview.NewTextView()

	initialText := getVisibleItems(root, "")
	textview.SetText(initialText)
	textview.ScrollToBeginning()
	textFilter := tview.NewInputField().
		SetLabel("Text Filter: ").
		SetFieldWidth(10)

	textFilter.SetChangedFunc(func(text string) {
        	textview.SetText(getVisibleItems(root, text))
	})

	flex := tview.NewFlex().
 		AddItem(textFilter, 0, 1, true).
 		AddItem(textview, 0, 12, false).
 		SetDirection(0)

        if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
        	panic(err)
        }
}
