package election

import (
	"fmt"
	"os"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"go.uber.org/zap"
)

// ElectionManager - handles the zookeeper election
type ElectionManager struct {
	zkConnection      *zk.Conn
	config            *Config
	isMaster          bool
	defaultACL        []zk.ACL
	logger            *zap.Logger
	electionChannel   chan int
	connectionChannel <-chan zk.Event
	messageChannel    chan int
	sessionID         int64
}

// New - creates a new instance
func New(config *Config, logger *zap.Logger, electionChannel chan int) (*ElectionManager, error) {

	return &ElectionManager{
		zkConnection:      nil,
		config:            config,
		defaultACL:        zk.WorldACL(zk.PermAll),
		logger:            logger,
		electionChannel:   electionChannel,
		messageChannel:    make(chan int),
		connectionChannel: nil,
	}, nil
}

// getNodeData - check if node exists
func (e *ElectionManager) getNodeData(node string) (*string, error) {

	data, _, err := e.zkConnection.Get(node)

	exists := true
	if err != nil {
		if err.Error() == "zk: node does not exist" {
			exists = false
		} else {
			return nil, err
		}
	}

	if !exists {
		return nil, nil
	}

	result := string(data)

	return &result, nil
}

// getZKMasterNode - returns zk master node name
func (e *ElectionManager) getZKMasterNode() (*string, error) {

	data, err := e.getNodeData(e.config.ZKElectionNodeURI)
	if err != nil {
		e.logError("getZKMasterNode", "error retrieving ZK election node data")
		return nil, err
	}

	return data, nil
}

// connect - connects to the zookeeper
func (e *ElectionManager) connect() error {

	e.logInfo("connect", "connecting to zookeeper...")

	var err error

	// Create the ZK connection
	e.zkConnection, e.connectionChannel, err = zk.Connect(e.config.ZKURL, time.Duration(e.config.SessionTimeout)*time.Second)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event := <-e.connectionChannel:
				if event.Type == zk.EventSession {
					if event.State == zk.StateConnected ||
						event.State == zk.StateConnectedReadOnly {
						e.logInfo("connect", "connection established with zookeeper")
					} else if event.State == zk.StateSaslAuthenticated ||
						event.State == zk.StateHasSession {
						e.logInfo("connect", "session created in zookeeper")
					} else if event.State == zk.StateAuthFailed ||
						event.State == zk.StateDisconnected ||
						event.State == zk.StateExpired {
						e.logInfo("connect", "zookeeper connection was lost")
						e.Close()
						e.messageChannel <- Disconnected
						for true {
							time.Sleep(time.Duration(e.config.ReconnectionTimeout) * time.Second)
							e.zkConnection, e.connectionChannel, err = zk.Connect(e.config.ZKURL, time.Duration(e.config.SessionTimeout)*time.Second)
							if err != nil {
								e.logError("connect", "error reconnecting to zookeeper: "+err.Error())
							} else {
								err := e.Start()
								if err != nil {
									e.logError("connect", "error starting election loop: "+err.Error())
								} else {
									break
								}
							}
						}
					}
				}
			}
		}
	}()

	return nil
}

// Start - starts to listen zk events
func (e *ElectionManager) Start() error {

	err := e.connect()
	if err != nil {
		e.logError("Start", "error connecting to zookeeper: "+err.Error())
		return err
	}

	err = e.electForMaster()
	if err != nil {
		e.logError("Start", "error electing this node for master: "+err.Error())
		return err
	}

	_, _, eventChannel, err := e.zkConnection.ExistsW(e.config.ZKElectionNodeURI)
	if err != nil {
		e.logError("Start", "error listening for zk events: "+err.Error())
		return err
	}

	go func() {
		for {
			select {
			case event := <-eventChannel:
				if event.Type == zk.EventNodeDeleted {
					e.logInfo("Start", "master has quit, trying to be the new master...")
					err := e.electForMaster()
					if err != nil {
						e.logError("Start", "error trying to elect this node for master: "+err.Error())
					}
				} else if event.Type == zk.EventNodeCreated {
					e.logInfo("Start", "a new master has been elected...")
				}
			case event := <-e.messageChannel:
				if event == Disconnected {
					e.logInfo("Start", "breaking election loop...")
					e.isMaster = false
					e.electionChannel <- Disconnected
					return
				}
			}
		}
	}()

	return nil
}

// Close - closes the connection
func (e *ElectionManager) Close() {

	if e.zkConnection != nil && !e.zkConnection.Disconnected() {
		e.zkConnection.Close()
	}

	time.Sleep(2 * time.Second)

	e.logInfo("Close", "ZK connection closed")
}

// getHostname - retrieves this node hostname from the OS
func (e *ElectionManager) getHostname() (string, error) {

	name, err := os.Hostname()
	if err != nil {
		e.logError("getHostname", "could not retrive this node hostname: "+err.Error())
		return "", err
	}

	return name, nil
}

// registerAsSlave - register this node as a slave
func (e *ElectionManager) registerAsSlave(nodeName string) error {

	data, err := e.getNodeData(e.config.ZKSlaveNodesURI)
	if err != nil {
		return err
	}

	if data == nil {
		path, err := e.zkConnection.Create(e.config.ZKSlaveNodesURI, []byte(nodeName), int32(0), e.defaultACL)
		if err != nil {
			e.logError("registerAsSlave", "error creating slave node directory: "+err.Error())
			return err
		}
		e.logInfo("registerAsSlave", "slave node directory created: "+path)
	}

	slaveNode := e.config.ZKSlaveNodesURI + "/" + nodeName

	data, err = e.getNodeData(slaveNode)
	if err != nil {
		return err
	}

	if data == nil {
		path, err := e.zkConnection.Create(slaveNode, []byte(nodeName), int32(zk.FlagEphemeral), e.defaultACL)
		if err != nil {
			e.logError("registerAsSlave", "error creating a slave node: "+err.Error())
			return err
		}

		e.logInfo("registerAsSlave", "slave node created: "+path)
	} else {
		e.logInfo("registerAsSlave", "slave node already exists: "+slaveNode)
	}

	e.isMaster = false
	e.electionChannel <- Slave

	return nil
}

// electForMaster - try to elect this node as the master
func (e *ElectionManager) electForMaster() error {

	name, err := e.getHostname()
	if err != nil {
		return err
	}

	zkMasterNode, err := e.getZKMasterNode()
	if err != nil {
		return err
	}

	if zkMasterNode != nil {
		if name == *zkMasterNode {
			e.logInfo("electForMaster", "this node is the master: "+*zkMasterNode)
			e.isMaster = true
		} else {
			e.logInfo("electForMaster", "another node is the master: "+*zkMasterNode)
			return e.registerAsSlave(name)
		}
	}

	path, err := e.zkConnection.Create(e.config.ZKElectionNodeURI, []byte(name), int32(zk.FlagEphemeral), e.defaultACL)
	if err != nil {
		if err.Error() == "zk: node already exists" {
			e.logInfo("electForMaster", "some node has became master before this node")
			return e.registerAsSlave(name)
		}

		e.logError("electForMaster", "error creating node: "+err.Error())
		return err
	}

	e.logInfo("electForMaster", "master node created: "+path)
	e.isMaster = true
	e.electionChannel <- Master

	slaveNode := e.config.ZKSlaveNodesURI + "/" + name
	slave, err := e.getNodeData(slaveNode)
	if err != nil {
		e.logError("electForMaster", fmt.Sprintf("error retrieving a slave node data '%s': %s\n", slaveNode, err.Error()))
		return nil
	}

	if slave != nil {
		err = e.zkConnection.Delete(slaveNode, 0)
		if err != nil {
			e.logError("electForMaster", fmt.Sprintf("error deleting slave node '%s': %s\n", slaveNode, err.Error()))
		} else {
			e.logInfo("electForMaster", "slave node deleted: "+slaveNode)
		}
	}

	return nil
}

// IsMaster - check if the cluster is the master
func (e *ElectionManager) IsMaster() bool {
	return e.isMaster
}

// GetClusterInfo - return cluster info
func (e *ElectionManager) GetClusterInfo() (*Cluster, error) {

	nodes := []string{}
	masterNode, err := e.getZKMasterNode()
	if err != nil {
		return nil, err
	}

	nodes = append(nodes, masterNode)

	slaveDir, err := e.getNodeData(e.config.ZKSlaveNodesURI)
	if err != nil {
		return nil, err
	}

	var children []string
	if slaveDir != nil {
		children, _, err = e.zkConnection.Children(e.config.ZKSlaveNodesURI)
		if err != nil {
			e.logError("GetClusterInfo", "error getting slave nodes: "+err.Error())
			return nil, err
		}

		nodes = append(nodes, children...)
	} else {
		children = []string{}
	}

	return &Cluster{
		IsMaster: e.isMaster,
		Master:   *masterNode,
		Slaves:   children,
		Nodes:    nodes,
		NumNodes: len(nodes),
	}, nil
}
