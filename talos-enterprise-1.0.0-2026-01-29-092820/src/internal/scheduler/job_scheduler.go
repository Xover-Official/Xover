package scheduler

import (
	"context"
	"sync"
	"time"
)

// Job represents a schedulable job
type Job struct {
	ID         string
	Name       string
	Type       JobType
	Payload    interface{}
	Priority   int // Higher = more important
	MaxRetries int
	RunAt      time.Time
	CreatedAt  time.Time
}

// JobType defines different job types
type JobType string

const (
	JobOptimization JobType = "optimization"
	JobResourceScan JobType = "resource_scan"
	JobCostAnalysis JobType = "cost_analysis"
	JobBackup       JobType = "backup"
	JobCleanup      JobType = "cleanup"
)

// JobHandler processes a job
type JobHandler func(context.Context, Job) error

// JobScheduler manages async job execution
type JobScheduler struct {
	mu       sync.RWMutex
	jobs     map[string]*Job
	queue    chan *Job
	handlers map[JobType]JobHandler
	workers  int
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(workers int, queueSize int) *JobScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	s := &JobScheduler{
		jobs:     make(map[string]*Job),
		queue:    make(chan *Job, queueSize),
		handlers: make(map[JobType]JobHandler),
		workers:  workers,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start workers
	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}

	// Start scheduler
	go s.scheduler()

	return s
}

// RegisterHandler registers a handler for a job type
func (s *JobScheduler) RegisterHandler(jobType JobType, handler JobHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[jobType] = handler
}

// Schedule schedules a job for execution
func (s *JobScheduler) Schedule(job Job) error {
	s.mu.Lock()
	job.CreatedAt = time.Now()
	if job.RunAt.IsZero() {
		job.RunAt = time.Now()
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = 3
	}
	s.jobs[job.ID] = &job
	s.mu.Unlock()

	return nil
}

// ScheduleDelayed schedules a job to run after a delay
func (s *JobScheduler) ScheduleDelayed(job Job, delay time.Duration) error {
	job.RunAt = time.Now().Add(delay)
	return s.Schedule(job)
}

// scheduler checks for due jobs and queues them
func (s *JobScheduler) scheduler() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.queueDueJobs()
		case <-s.ctx.Done():
			return
		}
	}
}

// queueDueJobs finds jobs that are due and queues them
func (s *JobScheduler) queueDueJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, job := range s.jobs {
		if job.RunAt.Before(now) || job.RunAt.Equal(now) {
			select {
			case s.queue <- job:
				delete(s.jobs, job.ID)
			default:
				// Queue full, try next time
			}
		}
	}
}

// worker processes jobs from the queue
func (s *JobScheduler) worker(id int) {
	defer s.wg.Done()

	for {
		select {
		case job := <-s.queue:
			s.executeJob(job)
		case <-s.ctx.Done():
			return
		}
	}
}

// executeJob executes a single job
func (s *JobScheduler) executeJob(job *Job) {
	s.mu.RLock()
	handler, exists := s.handlers[job.Type]
	s.mu.RUnlock()

	if !exists {
		return
	}

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Minute)
	defer cancel()

	err := handler(ctx, *job)

	if err != nil && job.MaxRetries > 0 {
		// Retry with exponential backoff
		job.MaxRetries--
		job.RunAt = time.Now().Add(time.Duration(3-job.MaxRetries) * time.Minute)
		s.Schedule(*job)
	}
}

// GetStatus returns the current scheduler status
func (s *JobScheduler) GetStatus() SchedulerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return SchedulerStatus{
		PendingJobs: len(s.jobs),
		QueuedJobs:  len(s.queue),
		Workers:     s.workers,
	}
}

// Close stops the scheduler
func (s *JobScheduler) Close() error {
	s.cancel()
	s.wg.Wait()
	close(s.queue)
	return nil
}

// SchedulerStatus represents scheduler state
type SchedulerStatus struct {
	PendingJobs int
	QueuedJobs  int
	Workers     int
}

// Helper to create jobs
func NewJob(id, name string, jobType JobType, payload interface{}) Job {
	return Job{
		ID:         id,
		Name:       name,
		Type:       jobType,
		Payload:    payload,
		Priority:   5,
		MaxRetries: 3,
		RunAt:      time.Now(),
	}
}
