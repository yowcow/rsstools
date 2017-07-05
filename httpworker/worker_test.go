package httpworker

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func httphandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ほげ"))
}

func TestWorker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(httphandler))
	defer server.Close()

	q := Queue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan *RssFeed),
		Out: make(chan *RssFeed),
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
				count += 1
				mx.Unlock()

				body, _ := ioutil.ReadAll(feed.Body)

				assert.Equal(t, true, feed.Attr["foo_flg"])
				assert.Equal(t, 1234, feed.Attr["bar_count"])
				assert.Equal(t, "ほげ", string(body))
			}
		}()
	}

	attr := RssAttr{
		"foo_flg":   true,
		"bar_count": 1234,
	}

	for i := 0; i < 20; i++ {
		feed := &RssFeed{
			Url:  server.URL,
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
