package loop

import (
	"fmt"
	"sync"
	"github.com/project-atlas/atlas/internal/cloud"
)

type ParallelExecutor struct {
	WorkerPoolSize int
}

func NewParallelExecutor(size int) *ParallelExecutor {
	return &ParallelExecutor{WorkerPoolSize: size}
}

func (pe *ParallelExecutor) BatchOptimize(resources []cloud.Resource, optimizeFn func(res cloud.Resource) error) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, pe.WorkerPoolSize)

	for _, res := range resources {
		wg.Add(1)
		go func(r cloud.Resource) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			if err := optimizeFn(r); err != nil {
				fmt.Printf("[ParallelExecutor] Error optimizing %s: %v\n", r.ID, err)
			}
		}(res)
	}

	wg.Wait()
}
