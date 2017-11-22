HTTP Get Request Worker
=======================

A worker that:
+ reads RSS feed object `*httpworker.RSSFeed` from channel `in`
+ makes HTTP GET request and appends response body `io.Reader` to the object
+ writes the object to channel `out`

HOW TO USE
----------

```go
in := make(chan *httpworker.RSSFeed)
q := httpworker.New(logger)
out := q.Start(in, 10) // boot 10 workers
```
