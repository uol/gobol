package election

// Config - configures the election
type Config struct {
	ZKURL             string
	ZKElectionNodeURI string
	ZKSlaveNodesURI   string
}

// Cluster - has cluster info
type Cluster struct {
	IsMaster bool
	Master   string
	Slaves   []string
}
