package httpworker

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/yowcow/rsstools/log"
)

var (
	UserAgent = "httpworker/1"
)

type RSSAttr map[string]interface{}

type RSSFeed struct {
	URL  string
	Attr RSSAttr
	Body *bytes.Buffer
}

type Queue struct {
	out    chan *RSSFeed
	logger log.Logger
	wg     *sync.WaitGroup
}

func New(logger log.Logger) *Queue {
	return &Queue{
		out:    make(chan *RSSFeed),
		logger: logger,
		wg:     new(sync.WaitGroup),
	}
}

func (q Queue) Start(in <-chan *RSSFeed, count int) <-chan *RSSFeed {
	q.wg.Add(count)
	for i := 1; i <= count; i++ {
		go q.runWorker(i, in)
	}
	return q.out
}

func (q Queue) Finish() {
	q.wg.Wait()
	close(q.out)
}

var createClient = func() *http.Client {
	return new(http.Client)
}

func (q Queue) runWorker(id int, in <-chan *RSSFeed) {
	defer func() {
		q.logger.Infof("[httpworker %d] Finished", id)
		q.wg.Done()
	}()

	q.logger.Infof("[httpworker %d] Started", id)
	client := createClient()

	for feed := range in {
		body, err := fetch(client, feed.URL)
		if err != nil {
			q.logger.Errorf("[httpworker %d] %s (%s)", id, err, feed.URL)
			continue
		}
		feed.Body = body
		q.out <- feed
	}
}

func fetch(client *http.Client, url string) (*bytes.Buffer, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("user-agent", UserAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected HTTP status 200 but got %d", resp.StatusCode)
	}

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
