package rssworker

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/yowcow/rsstools/httpworker"
)

var (
	Debug bool = false
)

type RssItem struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Attr  httpworker.RssAttr
}

type Rss1 struct {
	Items []*RssItem `xml:"item"`
}

type Rss2 struct {
	Channel *Rss1 `xml:"channel"`
}

type RssQueue struct {
	Wg  *sync.WaitGroup
	In  chan *httpworker.RssFeed
	Out chan *RssItem
}

func (self RssQueue) Start(id int) {
	defer self.Wg.Done()

	for feed := range self.In {
		var items []*RssItem
		rss1 := &Rss1{}
		rss2 := &Rss2{}

		rssXml, _ := ioutil.ReadAll(feed.Body)

		if err := xml.Unmarshal(rssXml, rss1); err != nil {
			fmt.Fprintf(os.Stdout, "[Rss Worker %d] Failed parsing XML %s\n", id, err)
			continue
		}

		if err := xml.Unmarshal(rssXml, rss2); err != nil {
			fmt.Fprintf(os.Stdout, "[Rss Worker %d] Failed parsing XML %s\n", id, err)
			continue
		}

		if len(rss1.Items) > 0 {
			items = rss1.Items
		} else if rss2.Channel != nil && len(rss2.Channel.Items) > 0 {
			items = rss2.Channel.Items
		}

		for _, item := range items {
			item.Attr = feed.Attr
			self.Out <- item
		}
	}
}
