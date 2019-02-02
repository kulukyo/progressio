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
Go and check examples/main.go.

## Run Tests
```shell
$ go test -v
```