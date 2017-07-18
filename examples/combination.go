package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/yowcow/rsstools/broadcaster"
	"github.com/yowcow/rsstools/httpworker"
	"github.com/yowcow/rsstools/itemworker"
	"github.com/yowcow/rsstools/rssworker"
)

var feedUrls = []string{
	"http://www3.nhk.or.jp/rss/news/cat0.xml",
	"https://news.yahoo.co.jp/pickup/rss.xml",
}

func httpQueue(workers int) httpworker.Queue {
	q := httpworker.Queue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan *httpworker.RssFeed),
		Out: make(chan *httpworker.RssFeed),
	}
	for i := 0; i < workers; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}
	return q
}

func rssQueue(workers int, in chan *httpworker.RssFeed) rssworker.Queue {
	q := rssworker.Queue{
		Wg:  &sync.WaitGroup{},
		In:  in,
		Out: make(chan *rssworker.RssItem),
	}
	for i := 0; i < workers; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}
	return q
}

func logQueue(workers int) itemworker.Queue {
	q := itemworker.Queue{
		Wg: &sync.WaitGroup{},
		In: make(chan *rssworker.RssItem),
		Task: func(item *rssworker.RssItem) bool {
			fmt.Fprintf(os.Stdout, "Link: %s, Title: %s\n", item.Link, item.Title)
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
		In: make(chan *rssworker.RssItem),
		Task: func(item *rssworker.RssItem) bool {
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

func broadcasterQueue(workers int, in chan *rssworker.RssItem, outs ...chan *rssworker.RssItem) broadcaster.Queue {
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

	httpQueue := httpQueue(4)
	rssQueue := rssQueue(4, httpQueue.Out)

	logQueue := logQueue(2)
	countQueue := countQueue(2, &count)

	broadcasterQueue := broadcasterQueue(2, rssQueue.Out, logQueue.In, countQueue.In)

	for _, url := range feedUrls {
		httpQueue.In <- &httpworker.RssFeed{url, nil, nil}
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
}
