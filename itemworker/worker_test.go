package itemworker

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yowcow/rsstools/rssworker"
)

func TestWorker_writes_to_out_chan(t *testing.T) {
	q := Queue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan *rssworker.RssItem),
		Out: make(chan *rssworker.RssItem),
		Task: func(item *rssworker.RssItem) bool {
			return true
		},
	}

	for i := 0; i < 4; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}

	wg := &sync.WaitGroup{}
	mx := &sync.Mutex{}
	count := 0

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _ = range q.Out {
				mx.Lock()
				count += 1
				mx.Unlock()
			}
		}()
	}

	for i := 0; i < 20; i++ {
		q.In <- &rssworker.RssItem{"Hoge", "http://hoge", nil}
	}
	close(q.In)
	q.Wg.Wait()

	close(q.Out)
	wg.Wait()

	assert.Equal(t, 20, count)
}

func TestWorker_no_write_to_out_chan(t *testing.T) {
	q := Queue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan *rssworker.RssItem),
		Out: make(chan *rssworker.RssItem),
		Task: func(item *rssworker.RssItem) bool {
			return false
		},
	}

	for i := 0; i < 4; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}

	wg := &sync.WaitGroup{}
	mx := &sync.Mutex{}
	count := 0

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _ = range q.Out {
				mx.Lock()
				count += 1
				mx.Unlock()
			}
		}()
	}

	for i := 0; i < 20; i++ {
		q.In <- &rssworker.RssItem{"Hoge", "http://hoge", nil}
	}
	close(q.In)
	q.Wg.Wait()

	close(q.Out)
	wg.Wait()

	assert.Equal(t, 0, count)
}
