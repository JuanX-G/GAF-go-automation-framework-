package scheduler

import (
	"container/heap"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type jobQueue []*queueEntry

func (q jobQueue) Len() int { return len(q) }

func (q jobQueue) Less(i, j int) bool {
	return q[i].nextRun.Before(q[j].nextRun)
}

func (q jobQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *jobQueue) Push(x any) {
	*q = append(*q, x.(*queueEntry))
}

func (q *jobQueue) Pop() any {
	old := *q
	n := len(old)
	x := old[n-1]
	*q = old[:n-1]
	return x
}

type JobFunc func(ctx context.Context)

type JobEntry struct {
	name  string
	spec  string
	time  time.Time
	cycle time.Duration
	job   JobFunc
}

type queueEntry struct {
	nextRun time.Time
	job     *JobEntry
}

type Scheduler struct {
	Sequential bool
	Mu         sync.Mutex
	Cond       *sync.Cond
	Jobs       map[string]*JobEntry
	queue      *jobQueue
}

func (s *Scheduler) Add(spec int, name string, jobArg JobFunc) error {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	if _, exists := s.Jobs[name]; exists {
		return fmt.Errorf("job of name %q already exists", name)
	}
	specS := strconv.Itoa(spec)

	td := time.Duration(spec * 1000000000)
	argTime := time.Now().Add(td)

	s.Jobs[name] = &JobEntry{
		name:  name,
		spec:  specS,
		time:  argTime,
		cycle: td,
		job:   jobArg,
	}
	// fmt.Println("job of name: ", name, "\nand parameters: ", s.Jobs[name], "created")
	return nil
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.Mu.Lock()
	q := &jobQueue{}
	heap.Init(q)
	for _, job := range s.Jobs {
		nextRun := time.Now().Add(job.cycle)
		heap.Push(q, &queueEntry{nextRun: nextRun, job: job})
	}
	s.queue = q
	s.Mu.Unlock()

	for {
		s.Mu.Lock()
		if s.queue.Len() == 0 {
			s.Mu.Unlock()
			select {
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		entry := (*s.queue)[0]
		wait := time.Until(entry.nextRun)
		s.Mu.Unlock()

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
			timer.Stop()
			jobCtx, cancel := context.WithTimeout(ctx, time.Minute)
			if !s.Sequential {
				go func(job *JobEntry) {
					defer cancel()
					job.job(jobCtx)
				}(entry.job)
			} else {
				entry.job.job(jobCtx)
				cancel()
			}

			s.Mu.Lock()
			heap.Pop(s.queue)
			entry.nextRun = time.Now().Add(entry.job.cycle)
			heap.Push(s.queue, entry)
			s.Mu.Unlock()

		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}
}

func (s *Scheduler) StartSeq(ctx context.Context) error {
	for {
		s.Mu.Lock()
		for s.queue.Len() == 0 {
			s.Cond.Wait()
		}
		entry := (*s.queue)[0]
		wait := time.Until(entry.nextRun)
		s.Mu.Unlock()

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
			timer.Stop()
			jobCtx, cancel := context.WithTimeout(ctx, time.Minute)
			entry.job.job(jobCtx)
			cancel()

			s.Mu.Lock()
			heap.Pop(s.queue)
			entry.nextRun = time.Now().Add(entry.job.cycle)
			heap.Push(s.queue, entry)
			s.Mu.Unlock()

		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		}
	}

}
