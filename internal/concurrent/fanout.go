package concurrent

import (
	"cloudsketch/internal/providers/azure/models"
	"log"
	"math"
	"sync"
)

type task struct {
	f func() ([]*models.Resource, error)
}

type result struct {
	resource []*models.Resource
	error    error
}

func FanOut(f []func() ([]*models.Resource, error)) ([]*models.Resource, error) {
	tasks := make(chan task, len(f))
	results := make(chan result, len(f))
	var wg sync.WaitGroup

	workers := int(math.Max(math.Sqrt(float64(len(f))), 1))

	log.Printf("fetch resources using %v workers", workers)

	// fan-out
	for range workers {
		wg.Add(1)
		go worker(tasks, results, &wg)
	}

	// send tasks
	go func() {
		for _, t := range f {
			tasks <- task{f: t}
		}

		close(tasks)
	}()

	// fan-in
	go func() {
		wg.Wait()
		close(results)
	}()

	res := []*models.Resource{}
	for result := range results {
		if result.error != nil {
			return nil, result.error
		}

		res = append(res, result.resource...)
	}

	return res, nil
}

func worker(tasks <-chan task, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		r, err := task.f()

		results <- result{
			resource: r,
			error:    err,
		}
	}
}
