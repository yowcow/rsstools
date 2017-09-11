package rssworker

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"github.com/yowcow/rsstools/httpworker"
)

var rssXML1 = `
<?xml version="1.0" encoding="UTF-8"?>
<rdf:RDF>
  <item>
    <title>あああ</title>
    <link>http://foobar</link>
  </item>
  <item>
    <title>いいい</title>
    <link>http://hogefuga</link>
  </item>
</rdf:RDF>
`
var rssXML2 = `
<?xml version="1.0" encoding="UTF-8"?>
<rdf:RDF>
  <channel>
    <item>
      <title>あああ</title>
      <link>http://foobar</link>
    </item>
    <item>
      <title>いいい</title>
      <link>http://hogefuga</link>
    </item>
  </channel>
</rdf:RDF>
`

func TestWorker_on_rss1(t *testing.T) {
	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	q := Queue{
		Wg:     &sync.WaitGroup{},
		In:     make(chan *httpworker.RSSFeed),
		Out:    make(chan *RSSItem),
		Logger: logger,
	}

	for i := 0; i < 4; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}

	result := map[string]int{}
	mx := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range q.Out {
				mx.Lock()
				result[item.Link]++
				mx.Unlock()

				assert.Equal(t, false, item.Attr["foo_flg"])
				assert.Equal(t, 1234, item.Attr["bar_count"])
			}
		}()
	}

	attr := httpworker.RSSAttr{
		"foo_flg":   false,
		"bar_count": 1234,
	}

	for i := 0; i < 10; i++ {
		buf := bytes.NewBufferString(rssXML1)
		feed := &httpworker.RSSFeed{"url", attr, buf}
		q.In <- feed
	}

	close(q.In)
	q.Wg.Wait()

	close(q.Out)
	wg.Wait()

	assert.Equal(t, 10, result["http://foobar"])
	assert.Equal(t, 10, result["http://hogefuga"])
	assert.Equal(t, 1, len(strings.Split(logbuf.String(), "\n")))
}

func TestWorker_on_rss2(t *testing.T) {
	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	q := Queue{
		Wg:     &sync.WaitGroup{},
		In:     make(chan *httpworker.RSSFeed),
		Out:    make(chan *RSSItem),
		Logger: logger,
	}

	for i := 0; i < 4; i++ {
		q.Wg.Add(1)
		go q.Start(i + 1)
	}

	result := map[string]int{}
	mx := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range q.Out {
				mx.Lock()
				result[item.Link]++
				mx.Unlock()

				assert.Equal(t, true, item.Attr["foo_flg"])
				assert.Equal(t, 1234, item.Attr["bar_count"])
			}
		}()
	}

	attr := httpworker.RSSAttr{
		"foo_flg":   true,
		"bar_count": 1234,
	}

	for i := 0; i < 10; i++ {
		buf := bytes.NewBufferString(rssXML2)
		feed := &httpworker.RSSFeed{"url", attr, buf}
		q.In <- feed
	}

	close(q.In)
	q.Wg.Wait()

	close(q.Out)
	wg.Wait()

	assert.Equal(t, 10, result["http://foobar"])
	assert.Equal(t, 10, result["http://hogefuga"])
	assert.Equal(t, 1, len(strings.Split(logbuf.String(), "\n")))
}

func TestWorker_on_invalid_xml(t *testing.T) {
	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	q := Queue{
		Wg:     &sync.WaitGroup{},
		In:     make(chan *httpworker.RSSFeed),
		Out:    make(chan *RSSItem),
		Logger: logger,
	}

	count := 0
	q.Wg.Add(1)
	go q.Start(1)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func(w *sync.WaitGroup, ch chan *RSSItem) {
		defer w.Done()
		for _ = range ch {
			count++
		}
	}(wg, q.Out)

	rssbuf := bytes.NewBufferString("something has happened")
	feed := &httpworker.RSSFeed{"http://something/rss", httpworker.RSSAttr{}, rssbuf}
	q.In <- feed
	q.In <- feed

	close(q.In)
	q.Wg.Wait()

	close(q.Out)
	wg.Wait()

	assert.Equal(t, 0, count)
	assert.Equal(t, 3, len(strings.Split(logbuf.String(), "\n")))
}
