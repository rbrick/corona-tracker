package main

import (
	"errors"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	errCompletedParsing = errors.New("completed parsing")
	pattern             = regexp.MustCompile("There are currently ([0-9]+,?[0-9]+) confirmed cases worldwide, including ([0-9]+,?[0-9]+) fatalities\\.")
)

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
	resp, err := http.Get(j.URL())
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
	return "https://docs.google.com/spreadsheets/d/1wQVypefm946ch4XDp37uZ-wartW4V7ILdg-qYiDXUHM/export?format=csv"
}

type BNONewsDataSource struct {
	records []*Record
}

func (b *BNONewsDataSource) Collect() (err error) {
	resp, err := http.Get(b.URL())

	if err != nil {
		return
	}

	node, err := html.Parse(resp.Body)

	if err != nil {
		return
	}

	var records []*Record

	walk(node, func(child *html.Node) error {
		if child.Type == html.ErrorNode {
			return io.EOF //
		} else if child.Type == html.ElementNode {
			if child.DataAtom == atom.P {

				if child.Parent.Type == html.ElementNode && child.Parent.DataAtom == atom.Div {

					mainContent := false
					for _, attr := range child.Parent.Attr {

						if attr.Key == "id" && attr.Val == "mvp-content-main" {
							// main content
							mainContent = true
							break
						}
					}

					if mainContent {
						if child.FirstChild.NextSibling != nil && child.FirstChild.NextSibling.DataAtom == atom.Strong {
							info := child.FirstChild.NextSibling.FirstChild.Data
							cases, _ := strconv.Atoi(strings.Replace(pattern.ReplaceAllString(info, "$1"), ",", "", -1))
							deaths, _ := strconv.Atoi(pattern.ReplaceAllString(info, "$2"))

							records = append(records, &Record{
								Province:       "",
								Country:        "Global",
								LastUpdated:    time.Now(),
								ConfirmedCases: cases,
								Deaths:         deaths,
								Recovered:      -1,
							})

							return errCompletedParsing
						}
					}
				}
			}
		}

		return nil
	})

	b.records = records
	return nil

}

func walk(node *html.Node, walkFunc func(child *html.Node) error) {
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		err := walkFunc(n)
		if err != nil {
			break
		}
		walk(n, walkFunc)
	}
}

func (b *BNONewsDataSource) Records() []*Record {
	return b.records
}

func (*BNONewsDataSource) URL() string {
	return "https://bnonews.com/index.php/2020/02/the-latest-coronavirus-cases/"
}
