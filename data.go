package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Data struct {
	XMLName xml.Name  `xml:"Root"`
	Items []Inventory `xml:"Inventory"`
}
type Inventory struct {
	XMLName  xml.Name `xml:"Inventory"`
	Category string   `xml:"Category"`
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
func getXMLData() *Data {
        tries := 0

	for tries < 31 {
                url := getDateURL(tries)
                xmlData, err := fetch(url)
                if err != nil {
                        tries += 1
                        continue
                }

                var data Data
        	err2 := xml.Unmarshal([]byte(xmlData), &data)
        	if err2 != nil {
                	tries += 1
                	continue
        	}

                return &data
	}

	panic("Could not find valid URL in 30 tries")
}

