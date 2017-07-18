package broadcaster

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yowcow/rsstools/itemworker"
	"github.com/yowcow/rsstools/rssworker"
)

func TestBroadcaster(t *testing.T) {
	q1Count := 0
	iq1 := itemworker.Queue{
		Wg: &sync.WaitGroup{},
		In: make(chan *rssworker.RSSItem),
		Task: func(item *rssworker.RSSItem) bool {
			q1Count++
			return false
		},
	}

	q2Count := 0
	iq2 := itemworker.Queue{
		Wg: &sync.WaitGroup{},
		In: make(chan *rssworker.RSSItem),
		Task: func(item *rssworker.RSSItem) bool {
			q2Count++
			return false
		},
	}

	iq1.Wg.Add(1)
	go iq1.Start(1)

	iq2.Wg.Add(1)
	go iq2.Start(1)

	bq := Queue{
		Wg:   &sync.WaitGroup{},
		In:   make(chan *rssworker.RSSItem),
		Outs: []chan *rssworker.RSSItem{iq1.In, iq2.In},
	}

	bq.Wg.Add(1)
	go bq.Start(1)

	for i := 0; i < 10; i++ {
		bq.In <- &rssworker.RSSItem{"hoge", "fuga", nil}
	}

	close(bq.In)
	bq.Wg.Wait()

	close(iq1.In)
	iq1.Wg.Wait()

	close(iq2.In)
	iq2.Wg.Wait()

	assert.Equal(t, 10, q1Count)
	assert.Equal(t, 10, q2Count)
}
