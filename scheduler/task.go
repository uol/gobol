package scheduler

import (
	"time"
)

//
// A single task to run a repetitive task
// author: rnojiri
//

// Job - a job to be executed
type job interface {
	Execute()
}

// Task - a scheduled task
type Task struct {
	ticker   *time.Ticker
	duration time.Duration
	job      job
	running  bool
}

// NewTask - creates a new task
func NewTask(duration time.Duration, job job) *Task {

	return &Task{
		duration: duration,
		job:      job,
		running:  false,
	}
}

// Start - starts to run this task
func (t *Task) Start() {

	if t.ticker != nil {
		t.ticker.Stop()
	}

	t.ticker = time.NewTicker(t.duration)

	go func() {
		for {
			<-t.ticker.C
			t.job.Execute()
		}
	}()

	t.running = true
}

// Stop - stops the task
func (t *Task) Stop() {

	t.ticker.Stop()
	t.running = false
}
