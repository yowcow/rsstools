package httpworker

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"sync"
)

var (
	UserAgent = "httpworker/1"
	Debug     = false
)

type HttpQueue struct {
	Wg  *sync.WaitGroup
	In  chan string
	Out chan *bufio.Reader
}

func (self HttpQueue) Start(id int) {
	defer self.Wg.Done()

	client := &http.Client{}

	for url := range self.In {
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Add("user-agent", UserAgent)
		resp, err := client.Do(req)

		if err != nil {
			if Debug {
				fmt.Fprintf(os.Stdout, "[Http Worker %d] %s (%s)\n", id, err, url)
			}
			self.Out <- nil
			continue
		}

		if Debug {
			fmt.Fprintf(os.Stdout, "[Http Worker %d] Status %d (%s)\n", id, resp.StatusCode, url)
		}

		r := bufio.NewReader(resp.Body)
		self.Out <- r
	}
}
