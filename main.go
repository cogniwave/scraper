package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Record struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

var MONTH_NAME_TO_EN = map[string]string{
	"Janeiro":   "January",
	"Fevereiro": "February",
	"Março":     "March",
	"Abril":     "April",
	"Maio":      "May",
	"Junho":     "June",
	"Julho":     "July",
	"Agosto":    "August",
	"Setembro":  "September",
	"Outubro":   "October",
	"Novembro":  "November",
	"Dezembro":  "December",
}

func isNewPost(date string, now time.Time) bool {
	parts := strings.Split(date, " / ")

	year, err := strconv.Atoi(parts[2])
	if err != nil {
		panic("failed to parse year")
	}

	if year < now.Year() {
		return false
	}

	if MONTH_NAME_TO_EN[parts[1]] != now.Month().String() {
		return false
	}

	day, err := strconv.Atoi(parts[0])
	if err != nil {
		panic("failed to parse day")
	}

	return day == now.Day()
}

func crawl(results chan Record) {
	collector := colly.NewCollector(colly.AllowedDomains("www.cargadetrabalhos.net", "cargadetrabalhos.net"))

	now := time.Now()
	abort := false

	collector.OnHTML("div.entrycontent", func(e *colly.HTMLElement) {
		if abort {
			return
		}

		e.ForEach("span.date", func(i int, h *colly.HTMLElement) {
			if !isNewPost(h.Text, now) {
				abort = true
			}
		})

		if abort {
			fmt.Println("aborting")
			return
		}

		e.ForEach("h2", func(i int, h *colly.HTMLElement) {
			if abort {
				fmt.Println("aborting")
				return
			}

			fmt.Println("got title", h.Text)

			results <- Record{
				Title: h.Text,
				Link:  h.ChildAttr("a", "href"),
			}
		})
	})

	collector.OnScraped(func(r *colly.Response) {
		fmt.Println("on scraped")
		close(results)
	})

	collector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	collector.Visit("https://www.cargadetrabalhos.net/category/web-design-programacao/")
}

func main() {
	results := make(chan Record)

	go crawl(results)

	for record := range results {
		fmt.Printf("Received record: %+v\n", record)
	}
}
