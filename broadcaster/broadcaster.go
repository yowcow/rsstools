package broadcaster

import (
	"sync"

	"github.com/yowcow/rsstools/rssworker"
)

type Queue struct {
	wg   *sync.WaitGroup
	outs []chan *rssworker.RSSItem
}

func New(count int) *Queue {
	outs := make([]chan *rssworker.RSSItem, count)
	for i := 0; i < len(outs); i++ {
		outs[i] = make(chan *rssworker.RSSItem)
	}
	return &Queue{
		wg:   new(sync.WaitGroup),
		outs: outs,
	}
}

func (q Queue) Start(in <-chan *rssworker.RSSItem, count int) []chan *rssworker.RSSItem {
	q.wg.Add(count)
	for i := 1; i <= count; i++ {
		go q.runWorker(i, in)
	}
	return q.outs
}

func (q Queue) Finish() {
	q.wg.Wait()
	for _, out := range q.outs {
		close(out)
	}
}

func (q Queue) runWorker(id int, in <-chan *rssworker.RSSItem) {
	defer q.wg.Done()
	for item := range in {
		for _, out := range q.outs {
			out <- item
		}
	}
}
