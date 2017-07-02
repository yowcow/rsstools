package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/yowcow/rsstools/httpworker"
	"github.com/yowcow/rsstools/itemworker"
	"github.com/yowcow/rsstools/rssworker"
)

var feedUrls = []string{
	"http://www3.nhk.or.jp/rss/news/cat0.xml",
	"https://news.yahoo.co.jp/pickup/rss.xml",
}

func main() {
	count := 0
	mx := &sync.Mutex{}

	httpQueue := httpworker.HttpQueue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan *httpworker.RssFeed),
		Out: make(chan *httpworker.RssFeed),
	}
	rssQueue := rssworker.RssQueue{
		Wg:  &sync.WaitGroup{},
		In:  httpQueue.Out,
		Out: make(chan *rssworker.RssItem),
	}
	logQueue := itemworker.ItemQueue{
		Wg:  &sync.WaitGroup{},
		In:  rssQueue.Out,
		Out: make(chan *rssworker.RssItem),
		Task: func(item *rssworker.RssItem) bool {
			fmt.Fprintf(os.Stdout, "Link: %s, Title: %s\n", item.Link, item.Title)
			return true
		},
	}
	countQueue := itemworker.ItemQueue{
		Wg: &sync.WaitGroup{},
		In: logQueue.Out,
		Task: func(item *rssworker.RssItem) bool {
			mx.Lock()
			defer mx.Unlock()
			count += 1
			return false
		},
	}

	for i := 0; i < 4; i++ {
		httpQueue.Wg.Add(1)
		go httpQueue.Start(i + 1)
	}

	for i := 0; i < 4; i++ {
		rssQueue.Wg.Add(1)
		go rssQueue.Start(i + 1)
	}

	for i := 0; i < 4; i++ {
		logQueue.Wg.Add(1)
		go logQueue.Start(i + 1)
	}

	for i := 0; i < 4; i++ {
		countQueue.Wg.Add(1)
		go countQueue.Start(i + 1)
	}

	for _, url := range feedUrls {
		httpQueue.In <- &httpworker.RssFeed{url, nil, nil}
	}

	close(httpQueue.In)
	httpQueue.Wg.Wait()

	close(rssQueue.In)
	rssQueue.Wg.Wait()

	close(logQueue.In)
	logQueue.Wg.Wait()

	close(countQueue.In)
	countQueue.Wg.Wait()

	fmt.Println("Fetched total =>", count)
}
