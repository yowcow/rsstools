package itemworker

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yowcow/rsstools/rssworker"
)

func TestWorker_writes_to_out_chan(t *testing.T) {
	type input struct {
		name string
		fn   RSSItemTask
	}
	type Case struct {
		input
		expectedCount int
		subtest       string
	}

	cases := []Case{
		{
			input{
				name: "always piped",
				fn: func(item *rssworker.RSSItem) bool {
					return true
				},
			},
			20,
			"task returns true pipes",
		},
		{
			input{
				name: "always not piped",
				fn: func(item *rssworker.RSSItem) bool {
					return false
				},
			},
			0,
			"task returns false is not piped",
		},
	}

	for _, c := range cases {
		t.Run(c.subtest, func(t *testing.T) {
			logbuf := new(bytes.Buffer)
			logger := log.New(logbuf, "", 0)

			in := make(chan *rssworker.RSSItem)
			q := New(c.input.name, c.input.fn, logger)
			out := q.Start(in, 4)

			count := 0
			done := make(chan bool)
			go func() {
				for _ = range out {
					count++
				}
				done <- true
			}()

			for i := 0; i < 20; i++ {
				in <- &rssworker.RSSItem{"Hoge", "http://hoge", nil}
			}

			close(in)
			q.Finish()
			<-done

			assert.Equal(t, c.expectedCount, count)
		})
	}
}
