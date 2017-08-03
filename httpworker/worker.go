package httpworker

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	Body *bytes.Buffer
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
				fmt.Fprintf(os.Stdout, "[HTTP Worker %d] %s (%s)\n", id, err, feed.URL)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stdout, "[HTTP Worker %d] Got error status %d (%s)\n", id, resp.StatusCode, feed.URL)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			fmt.Fprintf(os.Stdout, "[HTTP Worder %d] %s (%s)\n", id, err, feed.URL)
			continue
		}

		feed.Body = bytes.NewBuffer(body)
		q.Out <- feed

		resp.Body.Close()
	}
}
