package election

// Master - int signal for the election channel
const Master = 1

// Slave - int signal for the election channel
const Slave = 2

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
