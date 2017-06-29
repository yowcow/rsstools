package rssworker

import (
	"bufio"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var rssXml1 string = `
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
var rssXml2 string = `
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
	rssq := RssQueue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan *bufio.Reader),
		Out: make(chan *RssItem),
	}

	for i := 0; i < 4; i++ {
		rssq.Wg.Add(1)
		go rssq.Start(i + 1)
	}

	result := map[string]int{}
	mx := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range rssq.Out {
				mx.Lock()
				result[item.Link] += 1
				mx.Unlock()
			}
		}()
	}

	for i := 0; i < 10; i++ {
		r := bufio.NewReader(strings.NewReader(rssXml1))
		rssq.In <- r
	}

	close(rssq.In)
	rssq.Wg.Wait()

	close(rssq.Out)
	wg.Wait()

	assert.Equal(t, 10, result["http://foobar"])
	assert.Equal(t, 10, result["http://hogefuga"])
}

func TestWorker_on_rss2(t *testing.T) {
	rssq := RssQueue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan *bufio.Reader),
		Out: make(chan *RssItem),
	}

	for i := 0; i < 4; i++ {
		rssq.Wg.Add(1)
		go rssq.Start(i + 1)
	}

	result := map[string]int{}
	mx := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range rssq.Out {
				mx.Lock()
				result[item.Link] += 1
				mx.Unlock()
			}
		}()
	}

	for i := 0; i < 10; i++ {
		r := bufio.NewReader(strings.NewReader(rssXml2))
		rssq.In <- r
	}

	close(rssq.In)
	rssq.Wg.Wait()

	close(rssq.Out)
	wg.Wait()

	assert.Equal(t, 10, result["http://foobar"])
	assert.Equal(t, 10, result["http://hogefuga"])
}
