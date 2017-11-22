package main

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/labstack/gommon/log"
	"github.com/yowcow/rsstools/broadcaster"
	"github.com/yowcow/rsstools/httpworker"
	"github.com/yowcow/rsstools/itemworker"
	"github.com/yowcow/rsstools/rssworker"
)

var feedUrls = []string{
	"http://www3.nhk.or.jp/rss/news/cat0.xml",
	"https://news.yahoo.co.jp/pickup/rss.xml",
}

func main() {
	logbuf := new(bytes.Buffer)
	logger := log.New("")
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	// seed into http queue
	httpin := make(chan *httpworker.RSSFeed)

	// http worker queue
	httpqueue := httpworker.New(logger)
	httpout := httpqueue.Start(httpin, 4)

	// rss worker queue
	rssqueue := rssworker.New(logger)
	rssout := rssqueue.Start(httpout, 4)

	// broadcaster queue
	bcastqueue := broadcaster.New(2)
	bcastout := bcastqueue.Start(rssout, 4)

	// logging queue
	logqueue := itemworker.New("logging", func(item *rssworker.RSSItem) bool {
		logger.Infof("Link: %s, Title: %s", item.Link, item.Title)
		return false
	}, logger)
	logqueue.Start(bcastout[0], 4)

	// counting queue
	count := 0
	mx := new(sync.Mutex)
	countqueue := itemworker.New("counting", func(item *rssworker.RSSItem) bool {
		mx.Lock()
		defer mx.Unlock()
		count++
		return false
	}, logger)
	countqueue.Start(bcastout[1], 4)

	// now seed to cascade
	for _, url := range feedUrls {
		httpin <- &httpworker.RSSFeed{url, nil, nil}
	}

	// cascade close
	close(httpin)
	httpqueue.Finish()
	rssqueue.Finish()
	bcastqueue.Finish()
	logqueue.Finish()
	countqueue.Finish()

	fmt.Println("Fetched total =>", count)
	fmt.Println(logbuf.String())
}
