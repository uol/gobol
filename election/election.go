package election

import (
	"fmt"
	"os"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"go.uber.org/zap"
)

// Election - handles the zookeeper election
type Election struct {
	zkConnection *zk.Conn
	config       *Config
	isMaster     bool
	defaultACL   []zk.ACL
	logger       *zap.Logger
}

// New - creates a new instance
func New(config *Config, logger *zap.Logger) (*Election, error) {

	// Create the ZK connection
	zkConnection, _, err := zk.Connect([]string{config.ZKURL}, time.Second)
	if err != nil {
		return nil, err
	}

	e := &Election{
		zkConnection: zkConnection,
		config:       config,
		defaultACL:   zk.WorldACL(zk.PermAll),
		logger:       logger,
	}

	return e, nil
}

// getNodeData - check if node exists
func (e *Election) getNodeData(node string) (*string, error) {

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
func (e *Election) getZKMasterNode() (*string, error) {

	data, err := e.getNodeData(e.config.ZKElectionNodeURI)
	if err != nil {
		e.logError("getZKMasterNode", "error retrieving ZK election node data")
		return nil, err
	}

	return data, nil
}

// Start - starts to listen zk events
func (e *Election) Start() error {

	err := e.electForMaster()
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
			}
		}
	}()

	return nil
}

// Close - closes the connection
func (e *Election) Close() {

	e.zkConnection.Close()

	e.logInfo("Close", "ZK connection closed")
}

// getHostname - retrieves this node hostname from the OS
func (e *Election) getHostname() (string, error) {

	name, err := os.Hostname()
	if err != nil {
		e.logError("getHostname", "could not retrive this node hostname: "+err.Error())
		return "", err
	}

	return name, nil
}

// registerAsSlave - register this node as a slave
func (e *Election) registerAsSlave(nodeName string) error {

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

	return nil
}

// electForMaster - try to elect this node as the master
func (e *Election) electForMaster() error {

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
func (e *Election) IsMaster() bool {
	return e.isMaster
}

// GetClusterInfo - return cluster info
func (e *Election) GetClusterInfo() (*Cluster, error) {

	masterNode, err := e.getZKMasterNode()
	if err != nil {
		return nil, err
	}

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
	} else {
		children = []string{}
	}

	return &Cluster{
		IsMaster: e.isMaster,
		Master:   *masterNode,
		Slaves:   children,
	}, nil
}
