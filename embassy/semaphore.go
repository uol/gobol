package embassy

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

// Semaphore is an export of Consul's semaphore functionality
type Semaphore struct {
	*api.Semaphore
}

// NewSemaphore constructs a semaphore
func NewSemaphore(name, prefix string, limit int) (*Semaphore, error) {
	client := consulClient
	semaphore, err := client.SemaphoreOpts(&api.SemaphoreOptions{
		Prefix: prefix,
		Limit:  limit,

		SessionName: name,
	})
	return &Semaphore{
		Semaphore: semaphore,
	}, err
}

// NewResilientSemaphore constructs a semaphore that doesn't easily dies
func NewResilientSemaphore(name, prefix string, limit, ttl, retries int) (*Semaphore, error) {
	client := consulClient
	semaphore, err := client.SemaphoreOpts(&api.SemaphoreOptions{
		Prefix:           prefix,
		Limit:            limit,
		SessionName:      name,
		SessionTTL:       fmt.Sprintf("%dh", ttl),
		MonitorRetries:   retries,
		MonitorRetryTime: time.Hour * time.Duration(ttl) / 5,
		SemaphoreTryOnce: false,
	})
	return &Semaphore{
		Semaphore: semaphore,
	}, err
}
