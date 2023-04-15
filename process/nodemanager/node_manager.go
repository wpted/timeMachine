package nodemanager

import (
	"errors"

	"github.com/aarthikrao/timeMachine/components/dht"
	js "github.com/aarthikrao/timeMachine/components/jobstore"
	"github.com/aarthikrao/timeMachine/process/connectionmanager"
	dsm "github.com/aarthikrao/timeMachine/process/datastoremanager"
)

var (

	// If you face this err, it means that the nodeID was not found in
	ErrInvalidNodeIDSlotIDCombination = errors.New("invalid nodeid and slotid combination")

	// node has not yet been initalised
	ErrNotYetInitalised = errors.New("not yet initialised")
)

type NodeManager struct {
	selfNodeID string

	dsmgr *dsm.DataStoreManager

	connMgr *connectionmanager.ConnectionManager

	dhtMgr dht.DHT
}

func CreateNodeManager(
	selfNodeID string,
	dsmgr *dsm.DataStoreManager,
	connMgr *connectionmanager.ConnectionManager,
	dhtMgr dht.DHT,
) *NodeManager {
	return &NodeManager{
		selfNodeID: selfNodeID,
		dsmgr:      dsmgr,
		dhtMgr:     dhtMgr,
		connMgr:    connMgr,
	}
}

// Returns the location interface of the key. If the node is present on the same node,
// it returns the db, orelse it returns the connection to the respective server
func (nm *NodeManager) GetLocation(key string) (js.JobStore, error) {
	if nm.connMgr == nil {
		return nil, ErrNotYetInitalised
	}

	nodeVsSlot, err := nm.dhtMgr.GetLocation(key)
	if err != nil {
		return nil, err
	}

	// TODO: This is just a hack. Need to get the right algorithm based on leader and follower details
	presentInThisNode := false
	alternativeNodeID := ""
	for node, _ := range nodeVsSlot {
		if node == nm.selfNodeID {
			presentInThisNode = true
		} else {
			alternativeNodeID = node
		}
	}

	if presentInThisNode {
		slotNumber, ok := nodeVsSlot[nm.selfNodeID]
		if ok {
			return nm.dsmgr.GetDataNode(slotNumber)
		}
	}

	if alternativeNodeID != "" {
		return nm.connMgr.GetJobStore(nm.selfNodeID)
	}

	return nil, ErrInvalidNodeIDSlotIDCombination
}
