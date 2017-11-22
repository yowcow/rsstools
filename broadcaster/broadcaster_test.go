package broadcaster

import (
	"bytes"
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"github.com/yowcow/rsstools/itemworker"
	"github.com/yowcow/rsstools/rssworker"
)

func TestBroadcaster(t *testing.T) {
	logbuf := new(bytes.Buffer)
	logger := log.New("")
	//logger.SetLevel(log.ERROR)
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	in := make(chan *rssworker.RSSItem)
	bq := New(2)
	outs := bq.Start(in, 4)

	q1count := 0
	fn1 := func(item *rssworker.RSSItem) bool {
		q1count++
		return false
	}
	q1 := itemworker.New("q1", fn1, logger)
	q1.Start(outs[0], 4)

	q2count := 0
	fn2 := func(item *rssworker.RSSItem) bool {
		q2count++
		return false
	}
	q2 := itemworker.New("q2", fn2, logger)
	q2.Start(outs[1], 4)

	for i := 0; i < 10; i++ {
		in <- &rssworker.RSSItem{"hoge", "fuga", nil}
	}

	close(in)
	bq.Finish()
	q1.Finish()
	q2.Finish()

	assert.Equal(t, 10, q1count)
	assert.Equal(t, 10, q2count)
}
