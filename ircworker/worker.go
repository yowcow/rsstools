package ircworker

import (
	"fmt"
	"sync"

	"github.com/yowcow/rsstools/rssworker"
)

var (
	Debug = false
)

type Connection interface {
	Notice(string, string)
	Quit()
}

type IrcQueue struct {
	Wg   *sync.WaitGroup
	In   chan *rssworker.RssItem
	Out  chan *rssworker.RssItem
	Chan string
	Conn Connection
}

func (self IrcQueue) Start(id int) {
	defer func() {
		self.Wg.Done()
		self.Conn.Quit()
	}()

	for item := range self.In {
		self.Conn.Notice(self.Chan, fmt.Sprintf("%s %s", item.Title, item.Link))
		if self.Out != nil {
			self.Out <- item
		}
	}
}
