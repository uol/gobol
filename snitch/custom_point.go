package snitch

import (
	"sync"
)

//CustomPoint is the base structure used by Point, CustomPoint should be used if you need more control over your stats.
type CustomPoint struct {
	metric      string
	tags        map[string]string
	value       float64
	valNull     bool
	timestamp   int64
	count       int64
	interval    string
	aggregation string
	keepValue   bool
	sendOnNull  bool
	pre         PreTransform
	post        PostTransform
	sender      chan message

	mutex sync.Mutex
}

// ValueAdd adds v to value and increment the counter.
func (p *CustomPoint) ValueAdd(v float64) {
	p.mutex.Lock()
	p.value += v
	p.count++
	p.valNull = false
	p.mutex.Unlock()
}

// SetCount sets count to i.
func (p *CustomPoint) SetCount(i int64) {
	p.mutex.Lock()
	p.count = i
	p.mutex.Unlock()
}

// SetValue sets value to v.
func (p *CustomPoint) SetValue(v float64) {
	p.mutex.Lock()
	p.value = v
	p.valNull = false
	p.mutex.Unlock()
}

// SetTimestamp sets timestamp to t.
func (p *CustomPoint) SetTimestamp(t int64) {
	p.mutex.Lock()
	p.timestamp = t
	p.mutex.Unlock()
}

// GetValue return the current value.
func (p *CustomPoint) GetValue() float64 {
	return p.value
}

// GetCount return the current count.
func (p *CustomPoint) GetCount() int64 {
	return p.count
}

// GetTimestamp returns the current timestamp.
func (p *CustomPoint) GetTimestamp() int64 {
	return p.timestamp
}

// SetValueNull sets the value to zero and the valNull boolean to true.
func (p *CustomPoint) SetValueNull() {
	p.mutex.Lock()
	p.value = 0
	p.valNull = true
	p.mutex.Unlock()
}

// IsValueNull informs if the value is null or not.
func (p *CustomPoint) IsValueNull() bool {
	return p.valNull
}

// Run takes care of the logic for sending points
func (p *CustomPoint) Run() {
	send := p.pre(p)

	ts := p.timestamp
	if p.timestamp == 0 {
		ts = getTimeInMilliSeconds()
	}

	if send {
		p.sender <- message{
			Metric:    p.metric,
			Tags:      p.tags,
			Value:     p.GetValue(),
			Timestamp: ts,
		}
	}
	p.post(p)
}
