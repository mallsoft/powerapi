package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"text/template"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

type Zone struct {
	Name        string      `json:"name"`
	ZoneEntries []ZoneEntry `json:"data"`
}
type ZoneEntry struct {
	From  string  `json:"from_UTC"`
	To    string  `json:"to_UTC"`
	Price float64 `json:"price_NOK_KWh"`
}

func scrapeAllofThem() []Zone {
	fmt.Println("Scraping...")

	urls := urls()
	zc := make(chan Zone, len(urls))

	for k, v := range urls {

		go func(k string, v string) {
			zc <- scrapeZone(v, k)
		}(k, v)

	}

	var allZones []Zone
	for range urls {
		allZones = append(allZones, <-zc)
	}

	return allZones
}

func scrapeZone(target string, zone string) Zone {
	c := colly.NewCollector()
	extensions.RandomUserAgent(c)

	z := Zone{
		Name: zone,
	}

	c.OnHTML("tbody tr", func(e *colly.HTMLElement) {
		entry := ZoneEntry{}

		period := strings.Split(e.ChildText("td:nth-of-type(1)"), " - ")
		entry.From = period[0]
		entry.To = period[1]

		price, err := strconv.ParseFloat(e.ChildText("td:nth-of-type(2)"), 64)
		entry.Price = (price * ConversionRate) / 1000 // from MWh to kWh and convert to NOK

		if err != nil {
			fmt.Println("Price error ", err)
		} else {
			z.ZoneEntries = append(z.ZoneEntries, entry)
		}
	})

	c.OnResponse(func(r *colly.Response) {
		println(r.Request.Method, r.StatusCode, r.Request.URL.String())
	})

	c.Visit(target)

	return z
}

// returns a map of zones and their urls
func urls() map[string]string {
	now := time.Now().Format("02.01.2006")
	base := template.Must(template.New("url").Parse(`https://transparency.entsoe.eu/transmission-domain/r2/dayAheadPrices/show?name=&defaultValue=false&viewType=TABLE&areaType=BZN&atch=false&dateTime.dateTime={{.now}}+00%3A00%7CUTC%7CDAY&biddingZone.values={{.biddingZone}}&resolution.values=PT60M&dateTime.timezone=UTC`))
	zones := map[string]string{
		"NO1": "CTY%7C10YNO-0--------C!BZN%7C10YNO-1--------2",
		"NO2": "CTY%7C10YNO-0--------C!BZN%7C10YNO-2--------T",
		"NO3": "CTY%7C10YNO-0--------C!BZN%7C10YNO-3--------J",
		"NO4": "CTY%7C10YNO-0--------C!BZN%7C10YNO-4--------9",
		"NO5": "CTY%7C10YNO-0--------C!BZN%7C10Y1001A1001A48H",
	}

	for k, zone := range zones {
		b := bytes.NewBufferString("")
		err := base.Execute(b, map[string]interface{}{
			"now":         now,
			"biddingZone": zone,
		})
		if err != nil {
			panic(err)
		}

		zones[k] = b.String()
	}

	return zones
}
