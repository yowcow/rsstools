package httpworker

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"regexp"
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
	time.Sleep(time.Duration(TimeoutMillisecs+10) * time.Millisecond)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func TestWorkerSucceeds(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(httpOKHandler))
	defer server.Close()

	logbuf := new(bytes.Buffer)
	logger := log.New("")
	logger.SetLevel(log.ERROR)
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	in := make(chan *RSSFeed)
	q := New(logger)
	out := q.Start(in, 4)

	count := 0
	done := make(chan bool)
	go func(out <-chan *RSSFeed, count *int, done chan<- bool) {
		for feed := range out {
			*count++
			assert.Equal(t, true, feed.Attr["foo_flg"])
			assert.Equal(t, 1234, feed.Attr["bar_count"])
			assert.Equal(t, "ほげ", feed.Body.String())
		}
		done <- true
	}(out, &count, done)

	attr := RSSAttr{
		"foo_flg":   true,
		"bar_count": 1234,
	}
	for i := 0; i < 20; i++ {
		in <- &RSSFeed{server.URL, attr, nil}
	}

	close(in)
	q.Finish()
	<-done

	assert.Equal(t, 20, count)
	assert.Equal(t, "", logbuf.String())
}

func TestWorkerDoNothingOnRequestFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(httpErrorHandler))
	defer server.Close()

	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetLevel(log.ERROR)
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	in := make(chan *RSSFeed)
	q := New(logger)
	out := q.Start(in, 4)

	count := 0
	done := make(chan bool)
	go func(out <-chan *RSSFeed, count *int, done chan<- bool) {
		for _ = range out {
			*count++
		}
		done <- true
	}(out, &count, done)

	attr := RSSAttr{
		"foo_flg":   true,
		"bar_count": 1234,
	}
	feed := &RSSFeed{server.URL, attr, nil}

	in <- feed
	in <- feed

	close(in)
	q.Finish()
	<-done

	assert.Equal(t, 0, count)
	assert.Equal(t, 3, len(bytes.Split(logbuf.Bytes(), []byte("\n"))))
}

func TestWorkerDoNothingOnTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(httpTimeoutHandler))
	defer server.Close()

	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetLevel(log.ERROR)
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	createClient = func() *http.Client {
		client := new(http.Client)
		client.Timeout = time.Duration(TimeoutMillisecs) * time.Millisecond
		return client
	}

	in := make(chan *RSSFeed)
	q := New(logger)
	out := q.Start(in, 4)

	count := 0
	done := make(chan bool)
	go func(out <-chan *RSSFeed, count *int, done chan<- bool) {
		for _ = range out {
			*count++
		}
		done <- true
	}(out, &count, done)

	feed := &RSSFeed{
		URL:  server.URL,
		Attr: RSSAttr{},
	}

	in <- feed
	in <- feed

	close(in)
	q.Finish()
	<-done

	re := regexp.MustCompile("request canceled")

	assert.Equal(t, 0, count)
	assert.True(t, re.Match(logbuf.Bytes()))
}
