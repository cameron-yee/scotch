package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
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

type Target struct {
	url  string
	date *time.Time
}

type Result struct {
	target *Target
	data   *Data
}

func getTarget(startDate *time.Time, daysAgo int) *Target {
	targetDate := startDate.AddDate(0, 0, -daysAgo)
	formattedTargetDate := targetDate.Format("01_02_06")

	// https://www.dtnfspiritsandwines.com/AllLiquor_11_04_25.xml
	return &Target{url: fmt.Sprintf("https://www.dtnfspiritsandwines.com/AllLiquor_%s.xml", formattedTargetDate), date: &targetDate}
}

func fetchToChannel(target *Target, results chan<- *Result) {
	resp, err := http.Get(target.url)
	if err != nil {
		results <- &Result{target: target, data: nil}
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	xmlData := string(body)

	var data Data
	err = xml.Unmarshal([]byte(xmlData), &data)
	if err != nil {
		results <- &Result{target: target, data: nil}
		return
	}
	results <- &Result{target: target, data: &data}
	return
}

func getXMLDataInfos(startDate *time.Time) []*DataInfo {
	var targets []*Target

	tries := 0
	for tries < 365 {
		targets = append(targets, getTarget(startDate, tries))
		tries += 1
	}

	results := make(chan *Result, len(targets))
	for _, target := range targets {
		go fetchToChannel(target, results)
	}

	var dataInfos []*DataInfo
	var sortedResults []*Result
	for i := 0; i < len(targets); i++ {
		result := <-results
		sortedResults = append(sortedResults, result)
	}
	sort.Slice(sortedResults, func(i, j int) bool {
		return sortedResults[j].target.date.Before(*sortedResults[i].target.date)
	})
	for i := 0; i < len(sortedResults); i++ {
		result := sortedResults[i]
		if result.data != nil {
			dataInfos = append(dataInfos, &DataInfo{data: result.data, updatedDate: result.target.date})
		}
		if len(dataInfos) == 2 {
			return dataInfos
		}
	}
	panic("Could not find valid data in 365 tries.")
}

func getAppData() *AppData {
	today := time.Now()
	dataInfos := getXMLDataInfos(&today)

	return &AppData{
		latestData:   dataInfos[0],
		previousData: dataInfos[1],
	}
}
