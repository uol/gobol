package scheduler_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uol/gobol/scheduler"
)

//
// Tests for the scheduler package
// author: rnojiri
//

// IncJob - a job to increment it's counter
type IncJob struct {
	counter int
}

// Execute - increments the counter
func (ij *IncJob) Execute() {
	ij.counter++
}

// createScheduler - creates a new scheduler using 100 millis ticks
func createScheduler(taskId string, autoStart bool) (*IncJob, *scheduler.Manager) {

	job := &IncJob{}

	manager := scheduler.New()
	err := manager.AddTask(taskId, scheduler.NewTask(100*time.Millisecond, job), autoStart)
	if err != nil {
		panic(err)
	}

	return job, manager
}

// TestNoAutoStartTask - tests the scheduler with autostart feature disabled
func TestNoAutoStartTask(t *testing.T) {

	job, _ := createScheduler("x", false)

	time.Sleep(201 * time.Millisecond)

	assert.Equal(t, 0, job.counter, "expected no increment")
}

// TestAutoStartTask - tests the scheduler with autostart feature enabled
func TestAutoStartTask(t *testing.T) {

	job, _ := createScheduler("x", true)

	time.Sleep(101 * time.Millisecond)

	assert.Equal(t, 1, job.counter, "expected one increment")
}

// TestManualStartTask - tests the scheduler with manual start task
func TestManualStartTask(t *testing.T) {

	job, manager := createScheduler("x", false)

	time.Sleep(101 * time.Millisecond)

	assert.Equal(t, 0, job.counter, "expected no increment")

	manager.StartTask("x")

	time.Sleep(101 * time.Millisecond)

	assert.Equal(t, 1, job.counter, "expected one increment")
}

// TestStopTask - test the scheduler stop function
func TestStopTask(t *testing.T) {

	job, manager := createScheduler("x", true)

	time.Sleep(201 * time.Millisecond)

	manager.StopTask("x")

	time.Sleep(201 * time.Millisecond)

	assert.Equal(t, 2, job.counter, "expected two increments")
}

// TestRemoveTask - test the task removal function
func TestRemoveTask(t *testing.T) {

	job, manager := createScheduler("x", true)

	assert.Equal(t, 1, manager.GetNumTasks(), "expected only one task")

	time.Sleep(301 * time.Millisecond)

	assert.True(t, manager.RemoveTask("x"), "expected true")

	assert.Equal(t, 0, manager.GetNumTasks(), "expected no tasks")

	time.Sleep(301 * time.Millisecond)

	assert.Equal(t, 3, job.counter, "expected three increments")
}

// TestSimultaneousTasks - test multiple running tasks
func TestSimultaneousTasks(t *testing.T) {

	job1, manager := createScheduler("1", true)

	job2 := &IncJob{}
	manager.AddTask("2", scheduler.NewTask(50*time.Millisecond, job2), true)

	job3 := &IncJob{}
	manager.AddTask("3", scheduler.NewTask(200*time.Millisecond, job3), true)

	time.Sleep(401 * time.Millisecond)

	assert.Equal(t, 3, manager.GetNumTasks(), "expected three tasks")

	assert.Equal(t, 4, job1.counter, "expected four increments")
	assert.Equal(t, 8, job2.counter, "expected eigth increments")
	assert.Equal(t, 2, job3.counter, "expected two increments")

	assert.True(t, manager.RemoveTask("2"), "expected true")

	assert.Equal(t, 2, manager.GetNumTasks(), "expected three tasks")

	time.Sleep(401 * time.Millisecond)

	assert.Equal(t, 8, job2.counter, "expected eigth increments")
}

// TestRestartTask - test restarting a task after a while
func TestRestartTask(t *testing.T) {

	job, manager := createScheduler("x", true)

	time.Sleep(201 * time.Millisecond)

	assert.Equal(t, 2, job.counter, "expected two increments")

	if !assert.NoError(t, manager.StopTask("x"), "expected no error when stopping a task") {
		return
	}

	time.Sleep(201 * time.Millisecond)

	assert.Equal(t, 2, job.counter, "still expecting two increments")

	if !assert.NoError(t, manager.StartTask("x"), "expected no error when starting a task") {
		return
	}

	time.Sleep(201 * time.Millisecond)

	assert.Equal(t, 4, job.counter, "now expecting four increments")
}

// TestInexistentTask - test restarting a task after a while
func TestInexistentTask(t *testing.T) {

	_, manager := createScheduler("x", false)

	if !assert.False(t, manager.RemoveTask("y"), "expected 'false' when removing a non existing task") {
		return
	}

	if !assert.Error(t, manager.StartTask("y"), "expected 'error' when removing a non existing task") {
		return
	}
}
