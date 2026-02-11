package scheduler

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

type Job struct {
	id     	string
	runAt  	time.Time
	fn     	func()
	repeat 	RepeatPolicy
	index  	int
}

type RepeatPolicy interface {
    Next(after time.Time) time.Time
}

type Scheduler struct {
	mutex   sync.Mutex
	jobs  	jobHeap
	indexMap 	map[string]*Job
	timer 	*time.Timer
}

func New() *Scheduler {
	return &Scheduler{
		jobs:  jobHeap{},
		indexMap: make(map[string]*Job),
		timer: time.NewTimer(time.Hour),
	}
}

func (scheduler *Scheduler) Len() int {
	return scheduler.jobs.Len()
}

func (scheduler *Scheduler) AddAt(tm time.Time, fn func()) string {
	return scheduler.add(tm, fn, nil)
}

func (scheduler *Scheduler) AddAfter(duration time.Duration, fn func()) string {
	return scheduler.AddAt(time.Now().Add(duration), fn)
}

func (scheduler *Scheduler) Run(ctx context.Context) {
	for {
		select {
		case <-scheduler.timer.C:
			scheduler.fire()
		case <-ctx.Done():
			scheduler.timer.Stop()
			return
		}
	}
}

func (scheduler *Scheduler) Cancel(id string) bool {
	scheduler.mutex.Lock()
	defer scheduler.mutex.Unlock()

	job, ok := scheduler.indexMap[id]
	if !ok {
		return false
	}

	// remove job from heap
	heap.Remove(&scheduler.jobs, job.index)
	delete(scheduler.indexMap, id)
	
	scheduler.resetTimerLocked()

	return true
}

func (scheduler *Scheduler) PeekTime() (time.Time, bool) {
	scheduler.mutex.Lock()
	defer scheduler.mutex.Unlock()

	if len(scheduler.jobs) == 0 {
		return time.Time{}, false
	}

	return scheduler.jobs[0].runAt, true
}

func (scheduler *Scheduler) PeekID() (string, bool) {
	scheduler.mutex.Lock()
	defer scheduler.mutex.Unlock()

	if len(scheduler.jobs) == 0 {
		return "", false
	}

	return scheduler.jobs[0].id, true
}

func (scheduler *Scheduler) add(runAt time.Time, fn func(), repeat RepeatPolicy) string {
	scheduler.mutex.Lock()
	defer scheduler.mutex.Unlock()

	id := newID()
	job := &Job{
		id:     id,
		runAt:  runAt,
		fn:     fn,
		repeat: repeat,
	}

	heap.Push(&scheduler.jobs, job)

	scheduler.indexMap[id] = job
	scheduler.resetTimerLocked()

	return id
}

func (scheduler *Scheduler) fire() {
	scheduler.mutex.Lock()

	if len(scheduler.jobs) == 0 {
		scheduler.resetTimerLocked()

		scheduler.mutex.Unlock()
		return
	}

	// get top job
	job := heap.Pop(&scheduler.jobs).(*Job)
	delete(scheduler.indexMap, job.id)

	scheduler.mutex.Unlock()

	// run jon
	go job.fn()

	// prepare for repeat
	if job.repeat != nil {
		job.runAt = job.repeat.Next(job.runAt)

		scheduler.mutex.Lock()

		// add job back into queue
		heap.Push(&scheduler.jobs, job)

		// overwrite old job
		scheduler.indexMap[job.id] = job

		scheduler.resetTimerLocked()

		scheduler.mutex.Unlock()
	}
}


func (scheduler *Scheduler) resetTimerLocked() {
	if !scheduler.timer.Stop() {
		select {
		case <-scheduler.timer.C:
		default:
		}
	}

	// no jobs left, reset to default
	if len(scheduler.jobs) == 0 {
		scheduler.timer.Reset(time.Hour)
		return
	}

	// set timer to next runAt
	next := scheduler.jobs[0].runAt
	scheduler.timer.Reset(time.Until(next))
}

func newID() string {
	return time.Now().Format("20060102150405.000000000")
}