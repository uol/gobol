package cassandra

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/hailocab/go-hostpool"
)

// Settings defines the configuration for a cassandra cluster
type Settings struct {
	Keyspace     string
	Consistency  string
	Nodes        []string
	Username     string
	Password     string
	Connections  int
	Retry        int
	PageSize     int
	ProtoVersion int
	Timeout      string
}

// New creates a gocql.Session from a cassandra.Settings
func New(settings Settings) (*gocql.Session, error) {

	cluster := gocql.NewCluster(settings.Nodes...)

	if settings.Timeout != "" {
		t, err := time.ParseDuration(settings.Timeout)

		if err != nil {
			return nil, err
		}

		cluster.Timeout = t
	}

	cluster.Keyspace = settings.Keyspace

	cluster.ProtoVersion = 3

	if settings.ProtoVersion != 0 {
		cluster.ProtoVersion = settings.ProtoVersion
	}

	switch settings.Consistency {
	case "one":
		cluster.Consistency = gocql.One
	case "quorum":
		cluster.Consistency = gocql.Quorum
	case "all":
		cluster.Consistency = gocql.All
	}

	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: settings.Username,
		Password: settings.Password,
	}

	cluster.NumConns = settings.Connections

	cluster.RetryPolicy = &gocql.SimpleRetryPolicy{NumRetries: settings.Retry}

	//TokenAwarePolicy
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.HostPoolHostPolicy(hostpool.New(nil)))

	s, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	if settings.PageSize != 0 {
		s.SetPageSize(settings.PageSize)
	}

	return s, nil
}
