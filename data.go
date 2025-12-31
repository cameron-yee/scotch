package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type AppData struct {
	latestData   *DataInfo
	previousData *DataInfo
}

type DataInfo struct {
	data        *Data
	updatedDate *time.Time
}

type Data struct {
	XMLName xml.Name    `xml:"Root"`
	Items   []Inventory `xml:"Inventory"`
}

type Inventory struct {
	XMLName  xml.Name `xml:"Inventory"`
	Category string   `xml:"Category"`
	Item     string   `xml:"Item"`
	Price    float64  `xml:"Price"`
	UPC      int64    `xml:"UPC"`
}

func getDateURL(startDate *time.Time, daysAgo int) (string, *time.Time) {
	targetDate := startDate.AddDate(0, 0, -daysAgo)
	formattedTargetDate := targetDate.Format("01_02_06")

	// https://www.dtnfspiritsandwines.com/AllLiquor_11_04_25.xml
	return fmt.Sprintf("https://www.dtnfspiritsandwines.com/AllLiquor_%s.xml", formattedTargetDate), &targetDate
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

// Tries to get the correct URL with the latest date that has the XML data.
// Will try 30 times before giving up.
func getXMLData(startDate *time.Time) *DataInfo {
	tries := 0

	for tries < 366 {
		url, targetDate := getDateURL(startDate, tries)
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

		return &DataInfo{
			data:        &data,
			updatedDate: targetDate,
		}
	}

	panic(fmt.Sprintf("Could not find valid URL in 365 tries. startDate: %s", startDate))
}

func getAppData() *AppData {
	today := time.Now()
	latestData := getXMLData(&today)
	previousStartingDate := latestData.updatedDate.AddDate(0, 0, -1)
	previousData := getXMLData(&previousStartingDate)

	return &AppData{
		latestData:   latestData,
		previousData: previousData,
	}
}
