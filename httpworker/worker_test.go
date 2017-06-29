package httpworker

import (
	"bufio"
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

	httpq := HttpQueue{
		Wg:  &sync.WaitGroup{},
		In:  make(chan string),
		Out: make(chan *bufio.Reader),
	}

	for i := 0; i < 4; i++ {
		httpq.Wg.Add(1)
		go httpq.Start(i + 1)
	}

	count := 0
	mx := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _ = range httpq.Out {
				mx.Lock()
				count += 1
				mx.Unlock()
			}
		}()
	}

	for i := 0; i < 20; i++ {
		httpq.In <- server.URL
	}

	close(httpq.In)
	httpq.Wg.Wait()

	close(httpq.Out)
	wg.Wait()

	assert.Equal(t, 20, count)
}
