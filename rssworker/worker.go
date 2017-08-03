package rssworker

import (
	"encoding/xml"
	"fmt"
	"os"
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
	Wg  *sync.WaitGroup
	In  chan *httpworker.RSSFeed
	Out chan *RSSItem
}

func (q Queue) Start(id int) {
	defer q.Wg.Done()

	for feed := range q.In {
		var items []*RSSItem
		rss1 := &RSS1{}
		rss2 := &RSS2{}

		rssXML := feed.Body.Bytes()

		if err := xml.Unmarshal(rssXML, rss1); err != nil {
			fmt.Fprintf(os.Stdout, "[RSS Worker %d] Failed parsing XML %s\n", id, err)
			continue
		}

		if err := xml.Unmarshal(rssXML, rss2); err != nil {
			fmt.Fprintf(os.Stdout, "[RSS Worker %d] Failed parsing XML %s\n", id, err)
			continue
		}

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
