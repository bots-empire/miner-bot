package utils

import (
	"sync"
	"time"

	"github.com/Stepan1328/miner-bot/model"
)

type Spreader struct {
	// centre of all blocks
	blocks map[interface{}]*block

	// maximum downtime
	ttl time.Duration
	// mutex for running error when working with map
	mu *sync.Mutex
}

type block struct {
	// fifo processing queue
	pipe chan condition
	// last use time
	// used for delete idle blocks
	lastUse time.Time
}

type condition struct {
	handler      model.Handler
	situation    *model.Situation
	errorHandler func(error)
}

func (c condition) serve() error {
	return c.handler(c.situation)
}

func (b *block) serve() {
	go func() {
		for c := range b.pipe {
			b.lastUse = time.Now()

			if err := c.serve(); err != nil {
				c.errorHandler(err)
			}
		}
	}()

	return
}

// NewSpreader create new sort centre and start collector of old blocks
func NewSpreader(maxTTL time.Duration) *Spreader {
	s := &Spreader{
		blocks: make(map[interface{}]*block),

		ttl: maxTTL,
		mu:  &sync.Mutex{},
	}

	go s.collectObsoleteBlocks()

	return s
}

func (s *Spreader) collectObsoleteBlocks() {
	for now := range time.Tick(time.Second) {
		s.mu.Lock()
		for id, b := range s.blocks {
			if b.lastUse.Add(s.ttl).Before(now) {
				close(b.pipe)
				delete(s.blocks, id)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Spreader) ServeHandler(fn model.Handler, sit *model.Situation, errHandler func(error)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, ok := s.blocks[sit.User.ID]
	if !ok {
		b = &block{
			pipe: make(chan condition, 10),
		}
		b.serve()

		s.blocks[sit.User.ID] = b
	}

	b.pipe <- condition{
		handler:      fn,
		situation:    sit,
		errorHandler: errHandler,
	}
}
