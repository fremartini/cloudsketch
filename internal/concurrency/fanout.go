package concurrency

import (
	"log"
	"sync"
)

type task[T any] struct {
	f func() ([]T, error)
}

type result[T any] struct {
	value []T
	error error
}

func FanOut[T any](functions []func() ([]T, error)) ([]T, error) {
	tasks := make(chan task[T], len(functions))
	results := make(chan result[T], len(functions))
	var wg sync.WaitGroup

	workers := 2

	log.Printf("fetching resources using %v workers", workers)

	// fan-out
	for range workers {
		wg.Add(1)
		go worker(tasks, results, &wg)
	}

	// send tasks
	go func() {
		for _, t := range functions {
			tasks <- task[T]{f: t}
		}

		close(tasks)
	}()

	// fan-in
	go func() {
		wg.Wait()
		close(results)
	}()

	res := []T{}
	for result := range results {
		if result.error != nil {
			return nil, result.error
		}

		res = append(res, result.value...)
	}

	return res, nil
}

func worker[T any](tasks <-chan task[T], results chan<- result[T], wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		r, err := task.f()

		results <- result[T]{
			value: r,
			error: err,
		}
	}
}
