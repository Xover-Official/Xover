package loop

import (
	"fmt"
	"sync"

	"github.com/Xover-Official/Xover/internal/cloud"
)

type ParallelExecutor struct {
	WorkerPoolSize int
}

func NewParallelExecutor(size int) *ParallelExecutor {
	return &ParallelExecutor{WorkerPoolSize: size}
}

func (pe *ParallelExecutor) BatchOptimize(resources []*cloud.ResourceV2, optimizeFn func(res *cloud.ResourceV2) error) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, pe.WorkerPoolSize)

	for _, res := range resources {
		wg.Add(1)
		go func(r *cloud.ResourceV2) {
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
