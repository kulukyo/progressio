# progressio 
[![Build Status](https://travis-ci.com/kulukyo/progressio.svg?branch=master)](https://travis-ci.com/kulukyo/progressio) [![GoDoc](https://godoc.org/github.com/kulukyo/progressio?status.svg)](http://godoc.org/github.com/kulukyo/progressio)

Make progress of io.Reader and io.Writer observable.

## Features
- easy to use: NewProgressReader(...) or NewProgressWriter(...), and get progress reported!
- progress report while transferring bytes
  - progress in percentage
  - total transferred
  - transfer speed within a duration
- cancellation
- user defined report interval

## Installation
```shell
$ go get -u github.com/kulukyo/progressio
```

## Usage
```go
    // r is a normal io.Reader
    data := "some very looooooooooooooooooong data"
    r := strings.NewReader(data)

    // ticker for progress report
    t := time.NewTicker(100 * time.Millisecond)

    // wrap the normal reader "r" with context, data length
    // and a ticker, returns a wrapped reader and
    // a progress channel to receive progress
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    pr, progress := progressio.NewProgressReader(ctx, r, int64(len(data)), t)

    // consumes progress in another goroutine
    go func() {
       for p := range progress {
           fmt.Printf("reading: %v\n", p)
       }
    }()
    
    // use pr as a normal io.Reader,
    // for example in io.Copy, 
    io.Copy(ioutil.Discard, pr)
    // or pr.Read()
    //pr.Read(make([]byte, 20))
```
For detailed usage, go and check examples/main.go.

## Run Tests
```shell
$ go test -v
```
