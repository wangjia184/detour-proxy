package broker

import (
	"sync"
	"time"
)

type ConnectionSet struct {
	set   map[int64]*ConnectionInfo
	mutex sync.RWMutex
}

type ConnectionInfo struct {
	sourceNodeID string
	destNodeID   string
	timer        *time.Timer
}

func NewConnectionSet() *ConnectionSet {
	instance := &ConnectionSet{}
	instance.set = make(map[int64]*ConnectionInfo)
	instance.mutex = sync.RWMutex{}
	return instance
}

func (this *ConnectionSet) add(connID int64, sourceNodeID string, destNodeID string, timer *time.Timer) *ConnectionInfo {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	conn := &ConnectionInfo{
		sourceNodeID: sourceNodeID,
		destNodeID:   destNodeID,
		timer:        timer,
	}
	this.set[connID] = conn
	return conn
}

func (this *ConnectionSet) remove(connID int64) *ConnectionInfo {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	conn := this.set[connID]
	if conn != nil {
		delete(this.set, connID)
		return conn
	}
	return nil
}

func (this *ConnectionSet) get(connID int64) *ConnectionInfo {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.set[connID]
}
