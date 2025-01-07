package main

import (
	"context"
	"time"
)

type Client interface {
	Process(ctx context.Context, batch Batch) error
}

type countInTimeClient struct {
	s Service
	c chan struct{}
	t time.Time
	n uint64
}

func NewClient(s Service) Client {
	return countInTimeClient{
		s: s,
		c: make(chan struct{}, 1),
	}
}

func (c countInTimeClient) LockContext(ctx context.Context) error {
	select {
	case c.c <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c countInTimeClient) Unlock() {
	select {
	case <-c.c:
	default:
		panic("not locked")
	}
}

func (c countInTimeClient) Process(ctx context.Context, batch Batch) error {
	err := c.LockContext(ctx)
	if err != nil {
		return err
	}
	defer c.Unlock()

	n, p := c.s.GetLimits()

	for len(batch) > 0 {
		now := time.Now()
		if c.t.Before(now) {
			c.n = 0
			c.t = now.Add(p)
		}

		if c.n == n {
			t := time.NewTimer(c.t.Sub(now))
			select {
			case <-ctx.Done():
				t.Stop()
				return ctx.Err()
			case <-t.C:
				continue
			}
		}

		m := min(n-c.n, uint64(len(batch)))

		err = c.s.Process(ctx, batch[:m])
		if err != nil {
			return err
		}

		batch = batch[m:]
		c.n += m
	}

	return nil
}
