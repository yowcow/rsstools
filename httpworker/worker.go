package httpworker

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

var (
	UserAgent = "httpworker/1"
	Debug     = false
)

type RssAttr map[string]interface{}

type RssFeed struct {
	Url  string
	Attr RssAttr
	Body io.Reader
}

type Queue struct {
	Wg  *sync.WaitGroup
	In  chan *RssFeed
	Out chan *RssFeed
}

func (self Queue) Start(id int) {
	defer self.Wg.Done()

	client := &http.Client{}

	for feed := range self.In {
		req, _ := http.NewRequest("GET", feed.Url, nil)
		req.Header.Add("user-agent", UserAgent)
		resp, err := client.Do(req)

		if err != nil {
			if Debug {
				fmt.Fprintf(os.Stdout, "[Http Worker %d] %s (%s)\n", id, err, feed.Url)
			}
			self.Out <- nil
			continue
		}

		if Debug {
			fmt.Fprintf(os.Stdout, "[Http Worker %d] Status %d (%s)\n", id, resp.StatusCode, feed.Url)
		}

		feed.Body = bufio.NewReader(resp.Body)
		self.Out <- feed
	}
}
