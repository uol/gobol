package scheduler

import "fmt"

//
// Manages tasks to be executed repeatedly
// author: rnojiri
//

// Manager - schedules all expression executions
type Manager struct {
	taskMap map[string]*Task
}

// New - creates a new scheduler
func New() *Manager {

	return &Manager{
		taskMap: map[string]*Task{},
	}
}

// AddTask - adds a new task
func (m *Manager) AddTask(id string, task *Task, autoStart bool) error {

	if _, exists := m.taskMap[id]; exists {

		return fmt.Errorf("task id %s already exists", id)
	}

	m.taskMap[id] = task

	if autoStart {

		if task.running {
			return fmt.Errorf("task id %s already is running", id)
		}

		m.taskMap[id].Start()
	}

	return nil
}

// RemoveTask - removes a task
func (m *Manager) RemoveTask(id string) bool {

	if task, exists := m.taskMap[id]; exists {

		task.Stop()

		delete(m.taskMap, id)

		return true
	}

	return false
}

// StopTask - stops a task
func (m *Manager) StopTask(id string) error {

	if task, exists := m.taskMap[id]; exists {

		if task.running {
			task.Stop()
		} else {
			return fmt.Errorf("task id %s was not running (stop)", id)
		}

		return nil
	}

	return fmt.Errorf("task id %s does not exists (stop)", id)
}

// StartTask - starts a task
func (m *Manager) StartTask(id string) error {

	if task, exists := m.taskMap[id]; exists {

		if !task.running {
			task.Start()
		} else {
			return fmt.Errorf("task id %s is already running (start)", id)
		}

		return nil
	}

	return fmt.Errorf("task id %s does not exists (start)", id)
}

// GetNumTasks - returns the number of tasks
func (m *Manager) GetNumTasks() int {

	return len(m.taskMap)
}
