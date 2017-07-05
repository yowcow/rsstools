package itemworker

import (
	"sync"

	"github.com/yowcow/rsstools/rssworker"
)

type Queue struct {
	Wg   *sync.WaitGroup
	In   chan *rssworker.RssItem
	Out  chan *rssworker.RssItem
	Task func(*rssworker.RssItem) bool
}

func (self Queue) Start(id int) {
	defer self.Wg.Done()

	for item := range self.In {
		if self.Task(item) && self.Out != nil {
			self.Out <- item
		}
	}
}
