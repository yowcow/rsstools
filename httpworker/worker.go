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
	Wg     *sync.WaitGroup
	In     chan *RSSFeed
	Out    chan *RSSFeed
	client *http.Client
}

func (q Queue) Start(id int) {
	defer q.Wg.Done()
	q.client = &http.Client{}

	for feed := range q.In {
		buf, err := q.fetch(feed.URL)

		if err != nil {
			fmt.Fprintf(os.Stderr, "[HTTP Worker %d] %s (%s)\n", id, err, feed.URL)
			continue
		}

		feed.Body = buf

		q.Out <- feed
	}
}

func (q Queue) fetch(url string) (*bytes.Buffer, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", UserAgent)
	resp, err := q.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP status 200 but got %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(body), nil
}
