RSS Item Worker
===============

A worker that:
+ reads a RSS item object `*rssworker.RssItem` from channel `In`
+ calls a task registered as `Task`
+ writes the RSS item object to channel `Out` **if** the previous task returns `true`

HOW TO USE
----------

```go
itemQueue := rssworker.RssQueue{
    Wg:   &sync.WaitGroup{},
    In:   rssQueue.Out,
    Out:  make(chan *rssworker.RssItem),
    Task: func(*rssworker.RssItem) bool {
        // do whatever
        return true
    },
}

for i := 0; i < WORKER_COUNT; i++ {
    itemQueue.Wg.Add(1)
    go itemQueue.Start()
}
```
