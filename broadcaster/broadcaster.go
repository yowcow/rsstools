package broadcaster

import (
	"sync"

	"github.com/yowcow/rsstools/rssworker"
)

type Queue struct {
	Wg   *sync.WaitGroup
	In   chan *rssworker.RSSItem
	Outs []chan *rssworker.RSSItem
}

func (q Queue) Start(id int) {
	defer q.Wg.Done()

	for item := range q.In {
		for _, out := range q.Outs {
			out <- item
		}
	}
}
