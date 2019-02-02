package progressio_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kulukyo/progressio"
)

func newControlledTicker() (*time.Ticker, func()) {
	c := make(chan time.Time)
	ticker := &time.Ticker{
		C: c,
	}

	t := func() {
		c <- time.Now()
	}
	return ticker, t
}

type testrow struct {
	name string
	data string
	read []readstep
}

type readstep struct {
	buff     int
	n        int
	err      error
	cancel   bool
	expected progressio.Progress
}

func TestProgressReader(t *testing.T) {
	tt := []testrow{
		{name: "read step by step", data: "1234567890", read: []readstep{
			{buff: 1, n: 1, expected: progressio.Progress{Transferred: 1, Progress: float32(1) * 100 / 10}},
			{buff: 5, n: 5, expected: progressio.Progress{Transferred: 6, Progress: float32(6) * 100 / 10}},
			{buff: 8, n: 4, expected: progressio.Progress{Transferred: 10, Progress: float32(10) * 100 / 10}},
			{buff: 1, n: 0, err: io.EOF},
		}},
		{name: "read all", data: "1234567890", read: []readstep{
			{buff: 11, n: 10, expected: progressio.Progress{Transferred: 10, Progress: float32(10) * 100 / 10}},
			{buff: 1, n: 0, err: io.EOF},
		}},
		{name: "read empty", data: "1234567890", read: []readstep{
			{buff: 0, n: 0, expected: progressio.Progress{Transferred: 0, Progress: float32(0) * 100 / 10}},
		}},
		{name: "read after cancel", data: "1234567890", read: []readstep{
			{buff: 1, n: 0, cancel: true, err: context.Canceled},
		}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.data)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ticker, tick := newControlledTicker()
			pr, progress := progressio.NewProgressReader(ctx, r, int64(len(tc.data)), ticker)

			for _, step := range tc.read {
				if step.cancel {
					cancel()
				}
				buff := make([]byte, step.n)
				n, err := pr.Read(buff)

				if n != step.n {
					t.Errorf("n fail: expect %#v, got %#v", step.n, n)
				}
				if err != step.err {
					t.Errorf("error fail: expect %v, got %v", step.err, err)
				}

				if step.err != nil {
					continue
				}

				tick()
				p := <-progress
				if p.Transferred != step.expected.Transferred {
					t.Errorf("transferred fail: expect %#v, got %#v", step.expected.Transferred, p.Transferred)
				}
				if p.Progress != step.expected.Progress {
					t.Errorf("progress fail: expect %#v, got %#v", step.expected.Progress, p.Progress)
				}
			}
		})
	}
}

func TestProgressWriter(t *testing.T) {
	tt := []testrow{
		{name: "write step by step", data: "1234567890", read: []readstep{
			{buff: 1, n: 1, expected: progressio.Progress{Transferred: 1, Progress: float32(1) * 100 / 10}},
			{buff: 5, n: 5, expected: progressio.Progress{Transferred: 6, Progress: float32(6) * 100 / 10}},
			{buff: 4, n: 4, expected: progressio.Progress{Transferred: 10, Progress: float32(10) * 100 / 10}},
		}},
		{name: "write all", data: "1234567890", read: []readstep{
			{buff: 10, n: 10, expected: progressio.Progress{Transferred: 10, Progress: float32(10) * 100 / 10}},
		}},
		{name: "write empty", data: "1234567890", read: []readstep{
			{buff: 0, n: 0, expected: progressio.Progress{Transferred: 0, Progress: float32(0) * 100 / 10}},
		}},
		{name: "write after cancel", data: "1234567890", read: []readstep{
			{buff: 1, n: 0, cancel: true, err: context.Canceled},
		}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ticker, tick := newControlledTicker()
			pw, progress := progressio.NewProgressWriter(ctx, &b, int64(len(tc.data)), ticker)

			for _, step := range tc.read {
				if step.cancel {
					cancel()
				}
				buff := make([]byte, step.n)
				n, err := pw.Write(buff)

				if n != step.n {
					t.Errorf("n fail: expect %#v, got %#v", step.n, n)
				}
				if err != step.err {
					t.Errorf("error fail: expect %v, got %v", step.err, err)
				}

				if step.err != nil {
					continue
				}

				tick()
				p := <-progress
				if p.Transferred != step.expected.Transferred {
					t.Errorf("transferred fail: expect %#v, got %#v", step.expected.Transferred, p.Transferred)
				}
				if p.Progress != step.expected.Progress {
					t.Errorf("progress fail: expect %#v, got %#v", step.expected.Progress, p.Progress)
				}
			}
		})
	}
}

func TestProgressWriter_error(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := errorWriter{
		err: errors.New("some error"),
	}
	ticker, _ := newControlledTicker()
	pw, _ := progressio.NewProgressWriter(ctx, &w, 10, ticker)

	n, err := pw.Write(make([]byte, 1))
	if n != 0 || err != w.err {
		t.Errorf("error failed. expect %d %v, got %d %v", 0, w.err, n, err)
	}
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}
