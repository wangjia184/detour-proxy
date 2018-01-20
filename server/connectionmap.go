package server

import (
	"net"
	"sync"
)

type ConnectionMap struct {
	set   map[int64]net.Conn
	mutex sync.RWMutex
}

func NewConnectionMap() *ConnectionMap {
	instance := &ConnectionMap{}
	instance.set = make(map[int64]net.Conn)
	instance.mutex = sync.RWMutex{}
	return instance
}

func (this *ConnectionMap) add(connID int64, conn net.Conn) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	this.set[connID] = conn
}

func (this *ConnectionMap) get(connID int64) net.Conn {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.set[connID]
}

func (this *ConnectionMap) remove(connID int64) net.Conn {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	conn := this.set[connID]
	if conn != nil {
		delete(this.set, connID)
	}
	return conn
}
