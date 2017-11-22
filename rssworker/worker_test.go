package rssworker

import (
	"bytes"
	"strings"
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
	logbuf := new(bytes.Buffer)
	logger := log.New("")
	logger.SetLevel(log.ERROR)
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	in := make(chan *httpworker.RSSFeed)
	q := New(logger)
	out := q.Start(in, 4)

	result := map[string]int{}
	done := make(chan bool)
	go func() {
		for item := range out {
			result[item.Link]++
			assert.Equal(t, false, item.Attr["foo_flg"])
			assert.Equal(t, 1234, item.Attr["bar_count"])
		}
		done <- true
	}()

	attr := httpworker.RSSAttr{
		"foo_flg":   false,
		"bar_count": 1234,
	}
	for i := 0; i < 10; i++ {
		buf := bytes.NewBufferString(rssXML1)
		in <- &httpworker.RSSFeed{"url", attr, buf}
	}

	close(in)
	q.Finish()
	<-done

	assert.Equal(t, 10, result["http://foobar"])
	assert.Equal(t, 10, result["http://hogefuga"])
	assert.Equal(t, 1, len(strings.Split(logbuf.String(), "\n")))
}

func TestWorker_on_rss2(t *testing.T) {
	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetLevel(log.ERROR)
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	in := make(chan *httpworker.RSSFeed)
	q := New(logger)
	out := q.Start(in, 4)

	result := map[string]int{}
	done := make(chan bool)
	go func() {
		for item := range out {
			result[item.Link]++
			assert.Equal(t, true, item.Attr["foo_flg"])
			assert.Equal(t, 1234, item.Attr["bar_count"])
		}
		done <- true
	}()

	attr := httpworker.RSSAttr{
		"foo_flg":   true,
		"bar_count": 1234,
	}
	for i := 0; i < 10; i++ {
		buf := bytes.NewBufferString(rssXML2)
		in <- &httpworker.RSSFeed{"url", attr, buf}
	}

	close(in)
	q.Finish()
	<-done

	assert.Equal(t, 10, result["http://foobar"])
	assert.Equal(t, 10, result["http://hogefuga"])
	assert.Equal(t, 1, len(strings.Split(logbuf.String(), "\n")))
}

func TestWorker_on_invalid_xml(t *testing.T) {
	logbuf := &bytes.Buffer{}
	logger := log.New("")
	logger.SetLevel(log.ERROR)
	logger.SetOutput(logbuf)
	logger.SetHeader(`${level}`)

	in := make(chan *httpworker.RSSFeed)
	q := New(logger)
	out := q.Start(in, 4)

	count := 0
	done := make(chan bool)
	go func() {
		for _ = range out {
			count++
		}
		done <- true
	}()

	rssbuf := bytes.NewBufferString("something has happened")
	feed := &httpworker.RSSFeed{"http://something/rss", httpworker.RSSAttr{}, rssbuf}

	in <- feed
	in <- feed

	close(in)
	q.Finish()
	<-done

	assert.Equal(t, 0, count)
	assert.Equal(t, 3, len(strings.Split(logbuf.String(), "\n")))
}
