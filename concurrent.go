package concurrent

import (
	"go.uber.org/atomic"
)

type Pool struct {
	queue          chan interface{}
	worker         func(input interface{})
	totalThreads   atomic.Int64
	createdThreads atomic.Int64
	busyThreads    atomic.Int64
}

func New(threads int64, worker func(input interface{})) *Pool {
	if threads < 1 {
		threads = 1
	}

	pool := &Pool{
		queue:  make(chan interface{}),
		worker: worker,
	}
	pool.totalThreads.Store(threads)
	return pool
}

func (c *Pool) Process(input interface{}) {
	total := c.totalThreads.Load()
	busy := c.busyThreads.Load()
	created := c.createdThreads.Load()
	if busy == created && created < total {
		c.createdThreads.Inc()
		go func() {
			for {
				task, ok := <-c.queue
				if !ok {
					return
				}

				c.busyThreads.Inc()
				c.worker(task)
				c.busyThreads.Dec()
			}
		}()
	}

	c.queue <- input
}

func (c *Pool) Close() {
	close(c.queue)
}