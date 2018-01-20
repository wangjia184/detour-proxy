package broker

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"../dto"
	"github.com/gorilla/websocket"
)

type ProxyBroker struct {
	nodeSet       *NodeSet
	connectionSet *ConnectionSet
	upgrader      websocket.Upgrader
}

func Run(bindPort uint16) error {
	this := &ProxyBroker{}
	this.nodeSet = NewNodeSet()
	this.connectionSet = NewConnectionSet()
	this.upgrader = websocket.Upgrader{
		ReadBufferSize:    1024 * 1024,
		WriteBufferSize:   1024 * 1024,
		EnableCompression: true,
	}

	log.Println("Broker HTTP is listening on port", bindPort)
	httpServer := &http.Server{
		Addr:           fmt.Sprintf(":%d", bindPort),
		Handler:        this,
		ReadTimeout:    10 * time.Hour,
		WriteTimeout:   10 * time.Hour,
		MaxHeaderBytes: 1 << 20,
	}
	return httpServer.ListenAndServe()
}

func (this *ProxyBroker) ServeHTTP(writer http.ResponseWriter, req *http.Request) {

	log.Println(req.Method, req.URL)

	isServer := strings.EqualFold(req.URL.Query().Get("r"), "s")
	id := req.URL.Query().Get("id")

	//TODO: verify the request

	if len(id) > 0 {
		wsConn, err := this.upgrader.Upgrade(writer, req, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer wsConn.Close()

		node := this.nodeSet.add(id, isServer)
		defer (func() {
			this.nodeSet.remove(node)
		})()

		readerExitedChannel := make(chan bool)
		exited := false
		go (func() {
			defer (func() { exited = true })()

			for !exited {

				select {
				case buf, more := <-node.channel:
					{
						if more {
							// then send the chunk data
							err := wsConn.WriteMessage(websocket.BinaryMessage, buf)
							if err != nil {
								log.Println(id, err)
								return
							}
						} else {
							log.Println(id, "writer goroutine exited because channel was closed")
							return
						}
					} // case end
				case <-readerExitedChannel:
					{
						log.Println(id, "writer goroutine exited")
						return
					}
				case <-time.After(20 * time.Second):
					{
						// send an empty text message to keep connection alive
						err := wsConn.WriteMessage(websocket.TextMessage, nil)
						if err != nil {
							log.Println(id, err)
							return
						}
					} // case end
				} // select

			} //for {
		})()

		// reader
		for !exited {
			err := wsConn.SetReadDeadline(time.Now().Add(40 * time.Second))
			if err != nil {
				log.Println(err)
				break
			}

			mt, buffer, err := wsConn.ReadMessage()
			if err != nil {
				log.Println(err)
				break
			}

			if mt == websocket.BinaryMessage {

				err = this.handleInboundMessage(id, buffer)
				if err != nil {
					log.Println(err)
					break
				}
			}
		}

		log.Println(id, "reader goroutine exited")
	}
}

func (this *ProxyBroker) handleInboundMessage(nodeID string, buffer []byte) error {

	header, err := dto.DecodeHeader(buffer)
	if err != nil {
		return err
	}

	switch header.Type {
	case dto.Type_TCP_CONNECT:
		return this.handleConnecting(nodeID, header, buffer)

	case dto.Type_TCP_CONNECTION_ESTABLISHED:
		return this.handleConnectionStatusChange(nodeID, header, buffer)

	case dto.Type_TCP_CONNECTION_FAILED:
		return this.handleConnectionStatusChange(nodeID, header, buffer)

	case dto.Type_TCP_CONNECTION_CLOSED:
		return this.handleConnectionStatusChange(nodeID, header, buffer)

	case dto.Type_OUTBOUND_DATA:
		return this.handleData(nodeID, header, buffer)

	case dto.Type_INBOUND_DATA:
		return this.handleData(nodeID, header, buffer)

	default:
		log.Println("Unknown command type", header.Type)
		return nil
	}
}

func (this *ProxyBroker) handleConnecting(nodeID string, header *dto.MessageHeader, buffer []byte) error {

	srv := this.nodeSet.getServer()
	if srv == nil {
		srcChannel := this.nodeSet.getChannel(nodeID)
		if srcChannel == nil {
			return errors.New("Unable to found the source channel")
		}
		// no server to handle
		payload := &dto.Payload{
			ErrorMessage: "There is no server to handle connection at this moment",
		}
		bytes, err := dto.Encode(dto.Type_TCP_CONNECTION_FAILED, header.ConnectionID, payload)
		if err != nil {
			return err
		}

		srcChannel <- bytes
	} else {
		// timer to check if no response
		timer := time.AfterFunc(30*time.Second, func() {
			conn := this.connectionSet.remove(header.ConnectionID)
			if conn != nil {
				srcChannel := this.nodeSet.getChannel(nodeID)
				if srcChannel != nil {
					payload := &dto.Payload{
						ErrorMessage: fmt.Sprintf("Connection %v does not receive any reply from server after 30 seconds", header.ConnectionID),
					}
					bytes, err := dto.Encode(dto.Type_TCP_CONNECTION_FAILED, header.ConnectionID, payload)
					if err == nil {
						srcChannel <- bytes
					}
				}
			}
		})

		// record the connection
		this.connectionSet.add(header.ConnectionID, nodeID, srv.id, timer)
		srv.channel <- buffer
	}

	return nil
}

// connection is successfully established
func (this *ProxyBroker) handleConnectionStatusChange(nodeID string, header *dto.MessageHeader, buffer []byte) error {

	conn := this.connectionSet.get(header.ConnectionID)
	//log.Println(header.ConnectionID, header.Type, conn)
	if conn != nil {

		timer := conn.timer
		if timer != nil {
			timer.Stop()
			conn.timer = nil
		}

		// find the other end
		destNodeID := conn.sourceNodeID
		if destNodeID == nodeID {
			destNodeID = conn.destNodeID
		}
		srv := this.nodeSet.get(destNodeID)
		if srv != nil {
			srv.channel <- buffer
		} else {
			srcChannel := this.nodeSet.getChannel(nodeID)
			if srcChannel != nil {
				payload := &dto.Payload{
					ErrorMessage: fmt.Sprintf("Broker is unable to find the other end %v", destNodeID),
				}
				bytes, err := dto.Encode(dto.Type_TCP_CONNECTION_CLOSED, header.ConnectionID, payload)
				if err != nil {
					return err
				}
				srcChannel <- bytes
			}

		}

		if header.Type == dto.Type_TCP_CONNECTION_FAILED ||
			header.Type == dto.Type_TCP_CONNECTION_CLOSED {
			this.connectionSet.remove(header.ConnectionID)
		}
	} else if header.Type != dto.Type_TCP_CONNECTION_CLOSED {
		srcChannel := this.nodeSet.getChannel(nodeID)
		if srcChannel != nil {
			payload := &dto.Payload{
				ErrorMessage: fmt.Sprintf("Broker is unable to find the connection %v", header.ConnectionID),
			}
			bytes, err := dto.Encode(dto.Type_TCP_CONNECTION_CLOSED, header.ConnectionID, payload)
			if err != nil {
				return err
			}

			srcChannel <- bytes
		}
	}

	return nil
}

func (this *ProxyBroker) handleData(nodeID string, header *dto.MessageHeader, buffer []byte) error {
	conn := this.connectionSet.get(header.ConnectionID)
	if conn != nil {

		// find the other end
		destNodeID := conn.destNodeID
		if destNodeID == nodeID {
			destNodeID = conn.sourceNodeID
		}
		srv := this.nodeSet.get(destNodeID)
		if srv != nil {
			srv.channel <- buffer
			return nil
		}

		this.connectionSet.remove(header.ConnectionID)
	}

	srcChannel := this.nodeSet.getChannel(nodeID)
	if srcChannel != nil {
		payload := &dto.Payload{
			ErrorMessage: "The connection is lost",
		}
		bytes, err := dto.Encode(dto.Type_TCP_CONNECTION_CLOSED, header.ConnectionID, payload)
		if err == nil {
			srcChannel <- bytes
		}
	}
	return nil
}
