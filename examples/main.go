package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing/iotest"
	"time"

	"github.com/kulukyo/progressio"
)

func main() {
	data := "some very loooooooooooooooooooooong data"

	r := slowReader{
		Reader: iotest.OneByteReader(strings.NewReader(data)),
		sleep:  200 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	readTicker := time.NewTicker(100 * time.Millisecond)
	defer readTicker.Stop()
	pr, readProgress := progressio.NewProgressReader(ctx, r, int64(len(data)), readTicker)

	go func() {
		for p := range readProgress {
			fmt.Printf("reading: %v\n", p)
		}
	}()

	writeTicker := time.NewTicker(50 * time.Millisecond)
	defer writeTicker.Stop()
	pw, writeProgress := progressio.NewProgressWriter(ctx, ioutil.Discard, int64(len(data)), writeTicker)
	go func() {
		for p := range writeProgress {
			fmt.Printf("writing: %v\n", p)
		}
	}()

	io.Copy(pw, pr)
}

type slowReader struct {
	io.Reader
	sleep time.Duration
}

func (r slowReader) Read(p []byte) (int, error) {
	time.Sleep(r.sleep)
	n, err := r.Reader.Read(p)
	return n, err
}
