package itemworker

import (
	"sync"

	"github.com/yowcow/rsstools/rssworker"
)

type RSSItemTask func(*rssworker.RSSItem) bool

type Queue struct {
	Wg   *sync.WaitGroup
	In   chan *rssworker.RSSItem
	Out  chan *rssworker.RSSItem
	Task RSSItemTask
}

func (q Queue) Start(id int) {
	defer q.Wg.Done()

	for item := range q.In {
		if q.Task(item) && q.Out != nil {
			q.Out <- item
		}
	}
}
