package rssworker

import (
	"encoding/xml"
	"log"
	"sync"

	"github.com/yowcow/rsstools/httpworker"
)

var (
	Debug = false
)

type RSSItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Attr  httpworker.RSSAttr
}

type RSS1 struct {
	Items []*RSSItem `xml:"item"`
}

type RSS2 struct {
	Channel *RSS1 `xml:"channel"`
}

type Queue struct {
	Wg     *sync.WaitGroup
	In     chan *httpworker.RSSFeed
	Out    chan *RSSItem
	Logger *log.Logger
}

func (q Queue) Start(id int) {
	defer q.Wg.Done()

	for feed := range q.In {
		rssXML := feed.Body.Bytes()

		rss1, err := q.parseRSS1(rssXML)

		if err != nil {
			q.Logger.Printf("[RSS Worker %d] Failed parsing XML %s (%s)", id, err, feed.URL)
			continue
		}

		rss2, err := q.parseRSS2(rssXML)

		if err != nil {
			q.Logger.Printf("[RSS Worker %d] Failed parsing XML %s (%s)", id, err, feed.URL)
			continue
		}

		var items []*RSSItem

		if len(rss1.Items) > 0 {
			items = rss1.Items
		} else if rss2.Channel != nil && len(rss2.Channel.Items) > 0 {
			items = rss2.Channel.Items
		}

		for _, item := range items {
			item.Attr = feed.Attr
			q.Out <- item
		}
	}
}

func (q Queue) parseRSS1(rssXML []byte) (*RSS1, error) {
	rss1 := &RSS1{}

	if err := xml.Unmarshal(rssXML, rss1); err != nil {
		return nil, err
	}

	return rss1, nil
}

func (q Queue) parseRSS2(rssXML []byte) (*RSS2, error) {
	rss2 := &RSS2{}

	if err := xml.Unmarshal(rssXML, rss2); err != nil {
		return nil, err
	}

	return rss2, nil
}
