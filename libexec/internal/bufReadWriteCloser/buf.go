package bufReadWriteCloser

import (
	"bytes"
	"io"
	"sync"
	"sync/atomic"
)

type buffer struct {
	buffer  bytes.Buffer
	maxSize int
	mutex   sync.Mutex
	block   chan struct{} // block read when buffer is empty
	pool    sync.Pool
	close   int32
}

func New() io.ReadWriteCloser {
	const maxSize = 4 * 1024 * 1024 * 1024

	return &buffer{
		buffer:  bytes.Buffer{},
		maxSize: maxSize,
		block:   make(chan struct{}, 1),
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 8192)
			},
		},
	}
}

func (bf *buffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	for {
		bf.mutex.Lock()
		n, _ = bf.buffer.Read(p)
		bf.mutex.Unlock()

		if n > 0 {
			return n, nil
		}

		if atomic.LoadInt32(&bf.close) == 1 {
			return 0, io.EOF
		}

		<-bf.block
	}
}

func (bf *buffer) Write(p []byte) (n int, err error) {
	if atomic.LoadInt32(&bf.close) == 1 {
		return 0, Closed
	}

	size := len(p)

	bf.mutex.Lock()
	defer bf.mutex.Unlock()

	for bf.buffer.Len()+size > bf.maxSize {
		tmpB := bf.pool.Get().([]byte)
		_, _ = bf.buffer.Read(tmpB)
		bf.pool.Put(tmpB)
	}

	// cancel read block
	select {
	case bf.block <- struct{}{}:
	default:
	}

	return bf.buffer.Write(p)
}

func (bf *buffer) Close() error {
	if !atomic.CompareAndSwapInt32(&bf.close, 0, 1) {
		return nil
	}

	// cancel read block
	select {
	case bf.block <- struct{}{}:
	default:
	}

	return nil
}
