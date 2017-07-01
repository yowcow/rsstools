package ircworker

import (
	"fmt"
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yowcow/rsstools/rssworker"
)

type FakeConnection struct {
	Messages [][2]string
	Quitted  bool
}

func (conn *FakeConnection) Notice(target, input string) {
	conn.Messages = append(conn.Messages, [2]string{target, input})
}

func (conn *FakeConnection) Quit() {
	conn.Quitted = true
}

func TestWorker_write_to_irc_connection(t *testing.T) {
	conn := &FakeConnection{[][2]string{}, false}
	ircq := IrcQueue{
		Wg:   &sync.WaitGroup{},
		In:   make(chan *rssworker.RssItem),
		Chan: "#hoge",
		Conn: conn,
	}

	for i := 0; i < 4; i++ {
		ircq.Wg.Add(1)
		go ircq.Start(i + 1)
	}

	for i := 0; i < 20; i++ {
		ircq.In <- &rssworker.RssItem{
			Title: fmt.Sprintf("title%d", i),
			Link:  fmt.Sprintf("http://link/%d", i),
		}
	}

	close(ircq.In)
	ircq.Wg.Wait()

	re := regexp.MustCompile("^title\\d+ http://link/\\d+$")

	assert.True(t, conn.Quitted)
	assert.Equal(t, 20, len(conn.Messages))
	assert.Equal(t, "#hoge", conn.Messages[0][0])
	assert.True(t, re.MatchString(conn.Messages[0][1]))
}

func TestWorker_write_to_chan_if_given(t *testing.T) {
	conn := &FakeConnection{[][2]string{}, false}
	ircq := IrcQueue{
		Wg:   &sync.WaitGroup{},
		In:   make(chan *rssworker.RssItem),
		Out:  make(chan *rssworker.RssItem),
		Chan: "#hoge",
		Conn: conn,
	}

	for i := 0; i < 4; i++ {
		ircq.Wg.Add(1)
		go ircq.Start(i + 1)
	}

	wg := &sync.WaitGroup{}
	mx := &sync.Mutex{}
	count := 0
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _ = range ircq.Out {
				mx.Lock()
				count += 1
				mx.Unlock()
			}
		}()
	}

	for i := 0; i < 20; i++ {
		ircq.In <- &rssworker.RssItem{
			Title: fmt.Sprintf("title%d", i),
			Link:  fmt.Sprintf("http://link/%d", i),
		}
	}

	close(ircq.In)
	ircq.Wg.Wait()

	close(ircq.Out)
	wg.Wait()

	assert.Equal(t, 20, count)
}
