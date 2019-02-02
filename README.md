# progressio
Make progress of io.Reader and io.Writer observable.

## Features
- progress report while transferring bytes
  - progress in percentage
  - total transferred
  - transfer speed with a duration
- cancellation
- user defined report interval

## Installation
```shell
$ go get -u github.com/kulukyo/progressio
```

## Usage
```go
    // r is a normal io.Reader

    // ticker for progress report
    t := time.NewTicker(100 * time.Millisecond)

    // wrap a normal reader with context, original reader, 
    // data length and a ticker, return a reader 
    // and a progress channel to receive progress
    pr, progress := progressio.NewProgressReader(ctx.Background(), r, int64(len(data)), t)

    // consumes progress in another goroutine
    go func() {
       for p := range readProgress {
           fmt.Printf("reading: %v\n", p)
       }
    }()
    
    // use pr as normal io.Reader for example in io.Copy, 
    io.Copy(ioutil.Discard, pr)
    // or pr.Read()
    pr.Read(make([]byte, 20))
```
For detailed usage, go and check examples/main.go.

## Run Tests
```shell
$ go test -v
```