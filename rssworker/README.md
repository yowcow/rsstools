RSS Parsing Worker
==================

A worker that:
+ reads RSS feed object `*httpworker.RssFeed` from channel `In`
+ parses the RSS into `*rssworker.RssItem` objects
+ writes each item to channel `Out`

HOW TO USE
----------

```go
rssQueue := rssworker.RssQueue{
    Wg:  &sync.WaitGroup{},
    In:  httpQueue.Out,
    Out: make(chan *rssworker.RssItem),
}

for i := 0; i < WORKER_COUNT; i++ {
    rssQueue.Wg.Add(1)
    go rssQueue.Start()
}
```
