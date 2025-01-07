package main

import (
	"context"
	"fmt"
	"time"
)

type ExampleService struct{}

func (t ExampleService) GetLimits() (uint64, time.Duration) { return 2, time.Second }

func (t ExampleService) Process(context.Context, Batch) error { return nil }

func ExampleClient_Process() {
	s := ExampleService{}
	c := NewClient(s)
	t := time.Now()

	err := c.Process(context.TODO(), Batch{Item{}, Item{}, Item{}, Item{}, Item{}})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(time.Since(t).Round(time.Second).Seconds())

	// Output:
	// 2
}
