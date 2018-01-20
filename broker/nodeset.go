package broker

import (
	"sync"
)

type NodeSet struct {
	set     map[string]*Node
	servers []*Node
	mutex   sync.RWMutex
}

type Node struct {
	id       string
	channel  chan []byte
	isServer bool
}

func NewNodeSet() *NodeSet {
	instance := &NodeSet{}
	instance.servers = make([]*Node, 0, 100)
	instance.set = make(map[string]*Node)
	instance.mutex = sync.RWMutex{}
	return instance
}

func (this *NodeSet) add(id string, isServer bool) *Node {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	node := &Node{
		id:       id,
		isServer: isServer,
	}
	node.channel = make(chan []byte, 10)
	originalNode := this.set[id]
	this.set[id] = node
	if originalNode != nil {
		this.removeFromServerList(originalNode)
		if originalNode.channel != nil {
			close(originalNode.channel)
		}
	}

	if isServer {
		this.servers = append(this.servers, node)
	}
	return node
}

func (this *NodeSet) removeFromServerList(node *Node) {
	if node.isServer {
		deleted := 0
		for i := range this.servers {
			j := i - deleted
			if this.servers[j] == node {
				this.servers = this.servers[:j+copy(this.servers[j:], this.servers[j+1:])]
				deleted++
			}
		}
	}
}

func (this *NodeSet) remove(node *Node) {
	if node != nil {
		this.mutex.Lock()
		defer this.mutex.Unlock()

		currentNode := this.set[node.id]
		if currentNode == node {
			delete(this.set, node.id)
			this.removeFromServerList(node)
			if currentNode.channel != nil {
				close(currentNode.channel)
			}
		}
	}
}

func (this *NodeSet) get(id string) *Node {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.set[id]
}

func (this *NodeSet) getChannel(id string) chan []byte {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	node := this.set[id]
	if node != nil {
		return node.channel
	}
	return nil
}

func (this *NodeSet) getServer() *Node {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if len(this.servers) > 0 {
		return this.servers[0]
	}
	return nil
}
