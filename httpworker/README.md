HTTP Get Request Worker
=======================

Worker to read RSS feed object `*httpworker.RssFeed` from channel `In`, append response body `io.Reader` to the object, and writes the object to channel `Out`.

HOW TO USE
----------

```go
httpQueue := httpworker.HttpQueue{
    Wg:  &sync.WaitGroup{},
    In:  make(chan *httpworker.RssFeed),
    Out: make(chan *httpworker.RssFeed),
}

for i := 0; i < WORKER_COUNT; i++ {
    httpQueue.Wg.Add(1)
    go httpQueue.Start()
}
```
