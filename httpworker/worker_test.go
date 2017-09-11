package httpworker

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
)

var (
	TimeoutMillisecs = 100
)

func httpOKHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ほげ"))
}

func httpErrorHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Not Found"))
}

func httpTimeoutHandler(w http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Duration(TimeoutMillisecs) * time.Millisecond)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func TestWorkerSucceeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(httpOKHandler))
	defer server.Close()

	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	q := Queue{
		Wg:     &sync.WaitGroup{},
		In:     make(chan *RSSFeed),
		Out:    make(chan *RSSFeed),
		Logger: logger,
		Client: &http.Client{},
	}

	for i := 0; i < 4; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}

	count := 0
	mx := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for feed := range q.Out {
				mx.Lock()
				count++
				mx.Unlock()

				assert.Equal(t, true, feed.Attr["foo_flg"])
				assert.Equal(t, 1234, feed.Attr["bar_count"])
				assert.Equal(t, "ほげ", feed.Body.String())
			}
		}()
	}

	attr := RSSAttr{
		"foo_flg":   true,
		"bar_count": 1234,
	}

	for i := 0; i < 20; i++ {
		feed := &RSSFeed{
			URL:  server.URL,
			Attr: attr,
		}
		q.In <- feed
	}

	close(q.In)
	q.Wg.Wait()

	close(q.Out)
	wg.Wait()

	assert.Equal(t, 20, count)
	assert.Equal(t, "", logbuf.String())
	assert.Equal(t, 1, len(strings.Split(logbuf.String(), "\n")))
}

func TestWorkerDoNothingOnRequestFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(httpErrorHandler))
	defer server.Close()

	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	q := Queue{
		Wg:     &sync.WaitGroup{},
		In:     make(chan *RSSFeed),
		Out:    make(chan *RSSFeed),
		Logger: logger,
		Client: &http.Client{},
	}

	q.Wg.Add(1)
	go q.Start(1)

	count := 0
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func(ch chan *RSSFeed) {
		defer wg.Done()

		for _ = range ch {
			count++
		}
	}(q.Out)

	attr := RSSAttr{
		"foo_flg":   true,
		"bar_count": 1234,
	}

	feed := &RSSFeed{
		URL:  server.URL,
		Attr: attr,
	}
	q.In <- feed
	q.In <- feed

	close(q.In)
	q.Wg.Wait()

	close(q.Out)
	wg.Wait()

	assert.Equal(t, 0, count)
	assert.Equal(t, 3, len(strings.Split(logbuf.String(), "\n")))
}

func TestWorkerDoNothingOnTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(httpTimeoutHandler))
	defer server.Close()

	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	q := Queue{
		Wg:     &sync.WaitGroup{},
		In:     make(chan *RSSFeed),
		Out:    make(chan *RSSFeed),
		Logger: logger,
		Client: &http.Client{
			Timeout: time.Duration(TimeoutMillisecs-10) * time.Millisecond,
		},
	}

	q.Wg.Add(1)
	go q.Start(1)

	count := 0
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func(ch chan *RSSFeed) {
		defer wg.Done()

		for _ = range ch {
			count++
		}
	}(q.Out)

	feed := &RSSFeed{
		URL:  server.URL,
		Attr: RSSAttr{},
	}
	q.In <- feed

	close(q.In)
	q.Wg.Wait()

	close(q.Out)
	wg.Wait()

	re := regexp.MustCompile("request canceled")

	assert.Equal(t, 0, count)
	assert.True(t, re.Match(logbuf.Bytes()))
}
