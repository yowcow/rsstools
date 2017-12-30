package itemworker

import (
	"log"
	"sync"

	"github.com/yowcow/rsstools/rssworker"
)

type RSSItemTask func(*rssworker.RSSItem) bool

type Queue struct {
	name   string
	task   RSSItemTask
	wg     *sync.WaitGroup
	out    chan *rssworker.RSSItem
	logger *log.Logger
}

func New(name string, task RSSItemTask, logger *log.Logger) *Queue {
	return &Queue{
		name:   name,
		task:   task,
		wg:     new(sync.WaitGroup),
		out:    make(chan *rssworker.RSSItem),
		logger: logger,
	}
}

func (q Queue) Start(in <-chan *rssworker.RSSItem, count int) <-chan *rssworker.RSSItem {
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

func (q Queue) runWorker(id int, in <-chan *rssworker.RSSItem) {
	defer func() {
		q.logger.Printf("[itemworker %d (%s)] Finished", id, q.name)
		q.wg.Done()
	}()
	q.logger.Printf("[itemworker %d (%s)] Started", id, q.name)
	for item := range in {
		if q.task(item) {
			q.out <- item
		}
	}
}
