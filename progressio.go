package progressio

import (
	"context"
	"io"
	"time"
)

// Progress reports progress of transferring.
type Progress struct {
	Progress    float32       // progress from 0 ~ 100
	Speed       float32       // bytes per second
	Transferred int64         // total transferred bytes
	Duration    time.Duration // duration since last progress
}

// NewProgressReader wraps a reader reports progress and speed via channels while reading.
func NewProgressReader(ctx context.Context, r io.Reader, total int64, t *time.Ticker) (reader io.Reader, progress <-chan Progress) {
	p := &progressReader{
		Reader:     r,
		progresser: newProgresser(ctx, total, t),
	}

	p.work()

	return p, p.progress
}

// NewProgressWriter wraps a writer reports progress and speed via channels while writing.
func NewProgressWriter(ctx context.Context, w io.Writer, total int64, t *time.Ticker) (writer io.Writer, progress <-chan Progress) {
	p := &progressWriter{
		Writer:     w,
		progresser: newProgresser(ctx, total, t),
	}

	p.work()

	return p, p.progress
}

type progresser struct {
	ctx         context.Context
	total       int64
	ticker      *time.Ticker
	done        chan interface{}
	progress    chan Progress
	transferred chan int
}

func newProgresser(ctx context.Context, total int64, t *time.Ticker) progresser {
	return progresser{
		ctx:         ctx,
		total:       total,
		ticker:      t,
		transferred: make(chan int),
		done:        make(chan interface{}),
		progress:    make(chan Progress),
	}
}

func (p *progresser) work() {
	go func() {
		defer close(p.progress)
		defer close(p.done)
		defer close(p.transferred)
		totalTransferred, intervalTransferred := int64(0), 0
		begin := time.Now()
		for {
			select {
			case n := <-p.transferred:
				totalTransferred += int64(n)
				intervalTransferred += n
			case <-p.ticker.C:
				interval := time.Since(begin)
				begin = time.Now()
				p.progress <- Progress{
					Speed:       float32(float64(intervalTransferred) / interval.Seconds()),
					Progress:    float32((totalTransferred)*100) / float32(p.total),
					Transferred: totalTransferred,
					Duration:    interval,
				}
				intervalTransferred = 0
			case <-p.done:
				return
			case <-p.ctx.Done():
				return
			}
		}
	}()
}

type progressReader struct {
	io.Reader
	progresser
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	if pr.ctx.Err() != nil {
		return 0, pr.ctx.Err()
	}
	n, err = pr.Reader.Read(p)
	pr.transferred <- n
	if err != nil {
		pr.done <- true
	}
	return
}

type progressWriter struct {
	io.Writer
	progresser
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	if pw.ctx.Err() != nil {
		return 0, pw.ctx.Err()
	}
	n, err = pw.Writer.Write(p)
	pw.transferred <- n
	if err != nil {
		pw.done <- true
	}
	return
}
