package server

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/satori/go.uuid"

	"../comm"
	"../dto"
)

type ProxyServer struct {
	connections *ConnectionMap
	transport   comm.Transport
}

func Run(uri string) error {
	this := &ProxyServer{}
	this.connections = NewConnectionMap()

	if strings.LastIndex(uri, "?") > 0 {
		uri += "&"
	} else {
		uri += "?"
	}
	uri += "r=s&id=" + uuid.NewV4().String()

	transport, err := comm.NewWebSocketTransport(uri)
	if err != nil {
		panic(err)
	}
	this.transport = transport

	channel := make(chan dto.Message, 10) // this channel handles all input
	transport.RegisterChannel(0, channel)

	for {
		msg, more := <-channel
		if !more {
			log.Println("server.Run exits")
			break
		}

		switch msg.Header.Type {
		case dto.Type_TCP_CONNECT:
			go this.handleConnect(msg)

		case dto.Type_OUTBOUND_DATA:
			this.handleOutbound(msg)

		case dto.Type_TCP_CONNECTION_CLOSED:
			this.handleDisconnection(msg)

		default:
			log.Println("Unknown type:", msg.Header.Type)
		}
	}

	return nil
}

func (this *ProxyServer) handleConnect(msg dto.Message) {
	if msg.Payload == nil || msg.Header == nil {
		return
	}
	address := fmt.Sprintf("%v:%d", msg.Payload.Address, msg.Payload.Port)

	conn, err := net.DialTimeout("tcp", address, 20*time.Second)
	if err != nil {
		// handle error
		payload := &dto.Payload{
			ErrorMessage: err.Error(),
		}
		this.transport.Write(dto.Type_TCP_CONNECTION_FAILED, msg.Header.ConnectionID, payload)
		log.Println("Unable to dial", address, ",", err.Error())
		return
	}

	this.connections.add(msg.Header.ConnectionID, conn)

	// connected successfully
	this.transport.Write(dto.Type_TCP_CONNECTION_ESTABLISHED, msg.Header.ConnectionID, nil)
	//log.Println(msg.Header.ConnectionID, "connected")
	data := make([]byte, 1024*512, 1024*512)
	for {
		// receive the message
		n, err := conn.Read(data)
		if err != nil {
			payload := &dto.Payload{
				ErrorMessage: err.Error(),
			}
			this.transport.Write(dto.Type_TCP_CONNECTION_CLOSED, msg.Header.ConnectionID, payload)
			return
		}

		if n > 0 {
			payload := &dto.Payload{
				Data: data[0:n],
			}
			this.transport.Write(dto.Type_INBOUND_DATA, msg.Header.ConnectionID, payload)
		}

	}

}

func (this *ProxyServer) handleOutbound(msg dto.Message) {
	if msg.Header != nil {
		conn := this.connections.get(msg.Header.ConnectionID)
		if conn == nil { // connection has gone
			payload := &dto.Payload{
				ErrorMessage: fmt.Sprintf("Unable to find the connection whose id is %v", msg.Header.ConnectionID),
			}
			this.transport.Write(dto.Type_TCP_CONNECTION_CLOSED, msg.Header.ConnectionID, payload)
		} else { // forward the data
			_, err := conn.Write(msg.Payload.GetData())
			if err == nil {
				return
			}

			conn.Close()
			this.connections.remove(msg.Header.ConnectionID)
			payload := &dto.Payload{
				ErrorMessage: err.Error(),
			}
			this.transport.Write(dto.Type_TCP_CONNECTION_CLOSED, msg.Header.ConnectionID, payload)
		}
	}

}

func (this *ProxyServer) handleDisconnection(msg dto.Message) {
	if msg.Header != nil {
		conn := this.connections.remove(msg.Header.ConnectionID)
		if conn != nil {
			conn.Close()
		}
	}

}
