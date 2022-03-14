package sources

import (
	"context"
	"github.com/mmcdole/gofeed"
	"log"
	"time"
)

type RssParsedItem struct {
	ID    string
	Title string
	Text  string
	Links []string
	Time  time.Time
	Image string
	Meta  map[string]interface{}
}

func TestParse(url string) {
	checkTime := time.Now().Add(-time.Hour)
	ch := ParseRssFeed(context.Background(), url, &checkTime)
	for item := range ch {
		log.Printf("Test parse: %s (%s)\n", item.Title, item.Links[0])
	}
}

func NeedSkip(title string, links []string) bool {
	/*for _, link := range links {
		if strings.Contains(link, "/glamur/") {
			return true
		}
	}*/

	return false
}

func ParseRssFeed(ctx context.Context, url string, ignoreBefore *time.Time) <-chan RssParsedItem {
	fp := gofeed.NewParser()
	ch := make(chan RssParsedItem, 10)

	go func() {
		defer close(ch)
		log.Println("Started parsing " + url)

		feed, err := fp.ParseURL(url)
		if err != nil {
			log.Println("ParseRssFeed "+url+": ", err)
			return
		}

		for _, item := range feed.Items {
			result := RssParsedItem{
				Title: item.Title,
			}

			if item.UpdatedParsed != nil {
				result.Time = *item.UpdatedParsed
			} else if item.PublishedParsed != nil {
				result.Time = *item.PublishedParsed
			}

			if ignoreBefore != nil && result.Time.Before(*ignoreBefore) {
				continue
			}

			if item.GUID != "" {
				result.ID = item.GUID
			} else {
				result.ID = item.Link
			}

			if item.Content != "" {
				result.Text = item.Content
			} else if item.Description != "" {
				result.Text = item.Description
			}

			if item.Image != nil {
				result.Image = item.Image.URL
			}

			if len(item.Links) > 0 {
				for _, link := range item.Links {
					if link == item.Link {
						continue
					}

					result.Links = append(result.Links, link)
				}
			}
			result.Links = append(result.Links, item.Link)

			if NeedSkip(result.Title, result.Links) {
				continue
			}

			result.Meta = map[string]interface{}{
				"enclosures": item.Enclosures,
				"custom":     item.Custom,
			}

			ch <- result

			select {
			case <-ctx.Done():
				break
			default:
				continue
			}
		}
	}()

	return ch
}
