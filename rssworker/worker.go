package rssworker

import (
	"encoding/xml"
	"sync"

	"github.com/yowcow/rsstools/httpworker"
	"github.com/yowcow/rsstools/log"
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
	wg     *sync.WaitGroup
	out    chan *RSSItem
	logger log.Logger
}

func New(logger log.Logger) *Queue {
	return &Queue{
		wg:     new(sync.WaitGroup),
		out:    make(chan *RSSItem),
		logger: logger,
	}
}

func (q Queue) Start(in <-chan *httpworker.RSSFeed, count int) <-chan *RSSItem {
	q.wg.Add(count)
	for i := 1; i <= count; i++ {
		go q.runWorker(i, in)
	}
	return q.out
}

func (q Queue) Finish() {
	q.wg.Wait()
	close(q.out)
}

func (q Queue) runWorker(id int, in <-chan *httpworker.RSSFeed) {
	defer func() {
		q.logger.Infof("[rssworker %d] Finished", id)
		q.wg.Done()
	}()
	q.logger.Infof("[rssworker %d] Started", id)
	for feed := range in {
		rawxml := feed.Body.Bytes()

		rss1, err := parseRSS1(rawxml)
		if err != nil {
			q.logger.Errorf("[rssworker %d] Failed parsing XML as RSS1: %s (%s)", id, err, feed.URL)
			continue
		}

		rss2, err := parseRSS2(rawxml)
		if err != nil {
			q.logger.Errorf("[rssworker %d] Failed parsing XML as RSS2: %s (%s)", id, err, feed.URL)
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
			q.out <- item
		}
	}
}

func parseRSS1(rssXML []byte) (*RSS1, error) {
	rss1 := &RSS1{}
	if err := xml.Unmarshal(rssXML, rss1); err != nil {
		return nil, err
	}
	return rss1, nil
}

func parseRSS2(rssXML []byte) (*RSS2, error) {
	rss2 := &RSS2{}
	if err := xml.Unmarshal(rssXML, rss2); err != nil {
		return nil, err
	}
	return rss2, nil
}
