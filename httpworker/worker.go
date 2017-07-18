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

type RSSAttr map[string]interface{}

type RSSFeed struct {
	URL  string
	Attr RSSAttr
	Body io.Reader
}

type Queue struct {
	Wg  *sync.WaitGroup
	In  chan *RSSFeed
	Out chan *RSSFeed
}

func (q Queue) Start(id int) {
	defer q.Wg.Done()

	client := &http.Client{}

	for feed := range q.In {
		req, _ := http.NewRequest("GET", feed.URL, nil)
		req.Header.Add("user-agent", UserAgent)
		resp, err := client.Do(req)

		if err != nil {
			if Debug {
				fmt.Fprintf(os.Stdout, "[Http Worker %d] %s (%s)\n", id, err, feed.URL)
			}
			q.Out <- nil
			continue
		}

		if Debug {
			fmt.Fprintf(os.Stdout, "[Http Worker %d] Status %d (%s)\n", id, resp.StatusCode, feed.URL)
		}

		feed.Body = bufio.NewReader(resp.Body)
		q.Out <- feed
	}
}
