package main

import (
	"encoding/csv"
	"io"
	"strconv"
	"time"
)

type Record struct {
	Province, Country                 string
	LastUpdated                       time.Time
	ConfirmedCases, Deaths, Recovered int
}

type RecordDiff struct {
	// Added means it is a completely new record
	Added bool
	// Changes in the current data.
	DeltaCases, DeltaDeaths, DeltaRecovered int
}

func ReadRecords(reader io.Reader) []*Record {
	csvReader := csv.NewReader(reader)
	_records, _ := csvReader.ReadAll()

	var records []*Record

	for idx, v := range _records {
		if idx == 0 {
			continue
		}
		records = append(records, ParseRow(v))
	}

	return records
}

//DiffRecords implements a rather naive diff algorithm for determining new outbreaks
func DiffRecords(oldRecords, newRecords []*Record) []*RecordDiff {
	var diffs []*RecordDiff
	for _, newRecord := range newRecords {
		added := false
		for _, oldRecord := range oldRecords {
			if newRecord.Country == oldRecord.Country && oldRecord.Province == newRecord.Province {
				added = true

				diff := &RecordDiff{
					Added:          false,
					DeltaCases:     newRecord.ConfirmedCases - oldRecord.ConfirmedCases,
					DeltaDeaths:    newRecord.Deaths - oldRecord.Deaths,
					DeltaRecovered: newRecord.Recovered - oldRecord.Recovered,
				}

				diffs = append(diffs, diff)
				break
			}
		}

		if !added {
			// new record
			diffs = append(diffs, &RecordDiff{true, 0, 0, 0})
		}
	}
	return diffs
}

//ParseRow parses a row into a Record
func ParseRow(row []string) *Record {
	province := row[0]
	country := row[1]
	lastUpdate, _ := time.Parse("1/2/2006 15:04", row[2])

	cases := 0
	if row[3] != "" {
		cases, _ = strconv.Atoi(row[3])
	}
	deaths := 0
	if row[4] != "" {
		deaths, _ = strconv.Atoi(row[4])
	}
	recovered := 0
	if row[5] != "" {
		recovered, _ = strconv.Atoi(row[5])
	}

	return &Record{
		Province:       province,
		Country:        country,
		LastUpdated:    lastUpdate,
		ConfirmedCases: cases,
		Deaths:         deaths,
		Recovered:      recovered,
	}
}
