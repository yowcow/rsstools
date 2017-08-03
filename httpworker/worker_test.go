package httpworker

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func httpOKHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ほげ"))
}

func httpErrorHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "text/plain")
	w.WriteHeader(404)
	w.Write([]byte("Not Found"))
}

func TestWorkerSucceeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(httpOKHandler))
	defer server.Close()

	q := Queue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan *RSSFeed),
		Out: make(chan *RSSFeed),
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
}

func TestWorkerDoNothingOnRequestFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(httpErrorHandler))
	defer server.Close()

	q := Queue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan *RSSFeed),
		Out: make(chan *RSSFeed),
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

	q.In <- &RSSFeed{
		URL:  server.URL,
		Attr: attr,
	}

	close(q.In)
	q.Wg.Wait()

	close(q.Out)
	wg.Wait()

	assert.Equal(t, 0, count)
}
