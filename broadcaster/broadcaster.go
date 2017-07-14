package broadcaster

import (
	"sync"

	"github.com/yowcow/rsstools/rssworker"
)

type Queue struct {
	Wg   *sync.WaitGroup
	In   chan *rssworker.RssItem
	Outs []chan *rssworker.RssItem
}

func (self Queue) Start(id int) {
	defer self.Wg.Done()

	for item := range self.In {
		for _, out := range self.Outs {
			out <- item
		}
	}
}
