package logworker

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yowcow/rsstools/rssworker"
)

func TestWriteLog(t *testing.T) {
	buf := &bytes.Buffer{}
	w := bufio.NewWriter(buf)

	WriteLog(w, &rssworker.RssItem{
		Title: "ほげ",
		Link:  "http://hoge",
	})
	WriteLog(w, &rssworker.RssItem{
		Title: "ふが",
		Link:  "http://fuga",
	})
	w.Flush()

	data, _ := ioutil.ReadAll(buf)
	rows := strings.Split(string(data), "\n")

	assert.Equal(t, 3, len(rows))
}

func TestWorker_write_log(t *testing.T) {
	buf := &bytes.Buffer{}
	w := bufio.NewWriter(buf)

	logq := LogQueue{
		Wg:     &sync.WaitGroup{},
		In:     make(chan *rssworker.RssItem),
		Writer: w,
	}

	for i := 0; i < 4; i++ {
		logq.Wg.Add(1)
		go logq.Start(i + 1)
	}

	for i := 0; i < 20; i++ {
		logq.In <- &rssworker.RssItem{
			Title: fmt.Sprintf("ほげ%d", i),
			Link:  fmt.Sprintf("http://hoge/%d", i),
		}
	}

	close(logq.In)
	logq.Wg.Wait()

	w.Flush()
	rows := strings.Split(buf.String(), "\n")

	assert.Equal(t, 21, len(rows))
}

func TestWorker_write_to_chan_if_given(t *testing.T) {
	buf := &bytes.Buffer{}
	w := bufio.NewWriter(buf)

	logq := LogQueue{
		Wg:     &sync.WaitGroup{},
		In:     make(chan *rssworker.RssItem),
		Out:    make(chan *rssworker.RssItem),
		Writer: w,
	}

	for i := 0; i < 4; i++ {
		logq.Wg.Add(1)
		go logq.Start(i + 1)
	}

	wg := &sync.WaitGroup{}
	mx := &sync.Mutex{}
	count := 0

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _ = range logq.Out {
				mx.Lock()
				count += 1
				mx.Unlock()
			}
		}()
	}

	for i := 0; i < 20; i++ {
		logq.In <- &rssworker.RssItem{
			Title: fmt.Sprintf("ほげ%d", i),
			Link:  fmt.Sprintf("http://hoge/%d", i),
		}
	}

	close(logq.In)
	logq.Wg.Wait()

	close(logq.Out)
	wg.Wait()

	assert.Equal(t, 20, count)
}
