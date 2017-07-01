package logworker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/yowcow/rsstools/rssworker"
)

var (
	Debug   = false
	bufpool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
)

type LogQueue struct {
	Wg     *sync.WaitGroup
	In     chan *rssworker.RssItem
	Out    chan *rssworker.RssItem
	Writer io.Writer
}

func (self LogQueue) Start(id int) {
	defer self.Wg.Done()

	for item := range self.In {
		if Debug {
			fmt.Fprintf(os.Stdout, "[Log Worker %d] Got %s (%s)\n", id, item.Title, item.Link)
		}

		WriteLog(self.Writer, item)

		if self.Out != nil {
			self.Out <- item
		}
	}
}

func WriteLog(w io.Writer, item *rssworker.RssItem) {
	buf := bufpool.Get().(*bytes.Buffer)
	defer bufpool.Put(buf)

	buf.Reset()
	buf.WriteString(time.Now().Format(time.RFC3339Nano))
	buf.WriteByte(' ')
	buf.WriteString(item.Link)
	buf.WriteByte(' ')
	buf.WriteString(item.Title)
	buf.WriteString("\n")

	w.Write(buf.Bytes())
}
