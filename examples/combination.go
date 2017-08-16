package main

import (
	"bytes"
	"fmt"
	"net/http"
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

func httpQueue(workers int, logger *log.Logger) httpworker.Queue {
	q := httpworker.Queue{
		Wg:     &sync.WaitGroup{},
		In:     make(chan *httpworker.RSSFeed),
		Out:    make(chan *httpworker.RSSFeed),
		Logger: logger,
		Client: &http.Client{},
	}
	for i := 0; i < workers; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}
	return q
}

func rssQueue(workers int, in chan *httpworker.RSSFeed, logger *log.Logger) rssworker.Queue {
	q := rssworker.Queue{
		Wg:     &sync.WaitGroup{},
		In:     in,
		Out:    make(chan *rssworker.RSSItem),
		Logger: logger,
	}
	for i := 0; i < workers; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}
	return q
}

func logQueue(workers int, logger *log.Logger) itemworker.Queue {
	q := itemworker.Queue{
		Wg: &sync.WaitGroup{},
		In: make(chan *rssworker.RSSItem),
		Task: func(item *rssworker.RSSItem) bool {
			logger.Infof("Link: %s, Title: %s", item.Link, item.Title)
			return false
		},
	}
	for i := 0; i < workers; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}
	return q
}

func countQueue(workers int, count *int) itemworker.Queue {
	mx := &sync.Mutex{}
	q := itemworker.Queue{
		Wg: &sync.WaitGroup{},
		In: make(chan *rssworker.RSSItem),
		Task: func(item *rssworker.RSSItem) bool {
			mx.Lock()
			defer mx.Unlock()
			*count++
			return false
		},
	}
	for i := 0; i < workers; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}
	return q
}

func broadcasterQueue(workers int, in chan *rssworker.RSSItem, outs ...chan *rssworker.RSSItem) broadcaster.Queue {
	q := broadcaster.Queue{
		Wg:   &sync.WaitGroup{},
		In:   in,
		Outs: outs,
	}
	for i := 0; i < workers; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}
	return q
}

func main() {
	count := 0

	logbuf := bytes.Buffer{}
	logger := log.New("")
	logger.SetOutput(&logbuf)
	logger.SetHeader(`${level}`)

	httpQueue := httpQueue(4, logger)
	rssQueue := rssQueue(4, httpQueue.Out, logger)

	logQueue := logQueue(2, logger)
	countQueue := countQueue(2, &count)

	broadcasterQueue := broadcasterQueue(2, rssQueue.Out, logQueue.In, countQueue.In)

	for _, url := range feedUrls {
		httpQueue.In <- &httpworker.RSSFeed{url, nil, nil}
	}

	close(httpQueue.In)
	httpQueue.Wg.Wait()

	close(rssQueue.In)
	rssQueue.Wg.Wait()

	close(broadcasterQueue.In)
	broadcasterQueue.Wg.Wait()

	close(logQueue.In)
	logQueue.Wg.Wait()

	close(countQueue.In)
	countQueue.Wg.Wait()

	fmt.Println("Fetched total =>", count)
	fmt.Println(logbuf.String())
}
