package main

import "net/http"

//DataSource represents an remote source where we get our data & information this bot uses.
type DataSource interface {
	Collect() error
	//Records is the actual data we collected
	Records() []*Record
	//URL is the location of this data
	URL() string
}

type JohnsHopkinsSpreadsheetSource struct {
	records []*Record
}

func (j *JohnsHopkinsSpreadsheetSource) Collect() (err error) {
	resp, err := http.Get("https://docs.google.com/spreadsheets/d/1yZv9w9zRKwrGTaR-YzmAqMefw4wMlaXocejdxZaTs6w/export?format=csv")
	if err != nil {
		return
	}

	j.records = ReadRecords(resp.Body)
	return nil

}

func (j *JohnsHopkinsSpreadsheetSource) Records() []*Record {
	return j.records
}

func (j *JohnsHopkinsSpreadsheetSource) URL() string {
	return "https://docs.google.com/spreadsheets/d/1yZv9w9zRKwrGTaR-YzmAqMefw4wMlaXocejdxZaTs6w/export?format=csv"
}
