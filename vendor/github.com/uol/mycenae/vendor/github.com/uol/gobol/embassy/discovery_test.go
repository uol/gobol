package embassy

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func init() {
	var err error
	consulClient, err = api.NewClient(api.DefaultConfig())
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}

func TestServiceDiscovery(t *testing.T) {
	if output, err := exec.Command("pidof", "consul").CombinedOutput(); err != nil || len(output) == 0 {
		fmt.Fprintf(os.Stderr, "Consul was not found. Skipping tests\n")
		t.SkipNow()
	}
	any := false
	services, _, err := consulClient.Catalog().Services(nil)
	assert.NoError(t, err)
	for service := range services {
		port, err := GetServicePort(service, "http")

		if err == ErrNoServiceMatchesQuery {
			continue
		}

		assert.NoError(t, err)
		assert.NotZero(t, port)
		any = true
	}
	assert.True(t, any, "There was no affirmative answer from any server")
}
