package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"news/notifiers"
	"news/sources"
	html "news/utils"
	"os"
	"strings"
	"sync"
	"time"
)

/*
	Environment variables:

	TELEGRAM_TOKEN - Telegram Bot API token
	TARGET_USER_ID - User ID to send all news to
	RSS_LIST - File with RSS links (see rss-list-example.txt)
*/

func main() {
	debug := flag.Bool("debug", false, "Run in debug mode")
	silent := flag.Bool("silent", false, "Do not send anything via external APIs (Telegram etc)")
	earlier := flag.Bool("earlier", false, "Start parsing not from now, but 30 minutes ago (testing mode)")
	test := flag.String("test", "", "Test RSS parser using custom feed")
	flag.Parse()

	if *test != "" {
		if *test == "0" {
			rss := readRss()
			for _, s := range rss {
				sources.TestParse(s)
			}
		} else {
			sources.TestParse(*test)
		}
		return
	}

	if !(*debug) {
		notifiers.TestNotify(time.Now().Format(time.RFC1123 + "\nНовостной бот запущен"))
		defer func() {
			notifiers.TestNotify(time.Now().Format(time.RFC1123 + "\nНовостной бот остановлен"))
		}()
	}

	lastTime := time.Now()
	if *earlier {
		lastTime = lastTime.Add(-time.Hour)
	}
	var checkTime time.Time

	for {
		// Присвоить новое время нужно до начала парсинга, ведь вдруг за время, пока он происходит, появятся
		// ещё новости; для этого используется вспомогательная переменная со старым значением
		checkTime, lastTime = lastTime, time.Now()
		wg := &sync.WaitGroup{}

		// Обновляем список RSS каждую итерацию, вдруг с прошлого раза его поменяли
		rss := readRss()

		for _, s := range rss {
			wg.Add(1)

			go func(url string, checkTime time.Time, wg *sync.WaitGroup) {
				defer wg.Done()
				feedParsed := sources.ParseRssFeed(context.Background(), url, &checkTime)
				var text string

				for t := range feedParsed {
					text = ""
					if t.Text != "" {
						text = "\n\n" + html.StripHTML(t.Text)
					}

					if !*silent {
						log.Println("Parsed and ready to send: " + t.Links[0])
						newsHtml := fmt.Sprintf("<b>%s</b>%s\n\n<a href=\"%s\">%s</a>", html.StripHTML(t.Title), text, t.Links[0], t.Links[0])
						notifiers.TestNotify(newsHtml)
					} else {
						log.Printf("Parsed: %s (%s)\n", t.Title, t.Links[0])
					}
				}
			}(s, checkTime, wg)
		}

		wg.Wait()
		log.Println("Everything is parsed. Falling asleep...")
		time.Sleep(time.Minute * 5)
	}
}

func readRss() []string {
	filename := os.Getenv("RSS_LIST")
	if filename == "" {
		filename = "rss-list-example.txt"
	}

	var rss []string
	info, err := ioutil.ReadFile(filename)
	info = bytes.ReplaceAll(info, []byte("\r"), []byte(""))
	if err == nil {
		rss = strings.Split(string(info), "\n")
	}

	return rss
}
