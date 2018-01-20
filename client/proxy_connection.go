package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync/atomic"
	"time"

	"../comm"
	"../dto"
)

type ProxyConnection struct {
	transport      comm.Transport
	connectionId   int64
	channel        chan dto.Message
	remainingBytes []byte
}

var count uint32 = 0
var connectionIdBase int64 = int64(rand.New(rand.NewSource(time.Now().UnixNano())).Int31()) * 4294967296

func NewProxyConnection(address string, port uint16, transport comm.Transport) (*ProxyConnection, error) {
	instance := &ProxyConnection{
		transport: transport,
	}

	// generate a unique connection id
	seq := atomic.AddUint32(&count, 1)
	instance.connectionId = connectionIdBase + int64(seq)

	// construct the payload
	payload := &dto.Payload{
		Address: address,
		Port:    int32(port),
	}

	instance.channel = make(chan dto.Message, 10)
	transport.RegisterChannel(instance.connectionId, instance.channel)

	// send the message
	err := transport.Write(dto.Type_TCP_CONNECT, instance.connectionId, payload)
	if err != nil {
		transport.UnregisterChannel(instance.connectionId, instance.channel)
		return nil, err
	}

	// receive the message
	select {
	case msg := <-instance.channel:
		if msg.Header.Type == dto.Type_TCP_CONNECTION_FAILED {
			transport.UnregisterChannel(instance.connectionId, instance.channel)
			if msg.Payload != nil && len(msg.Payload.ErrorMessage) > 0 {
				return nil, errors.New(msg.Payload.ErrorMessage)
			}
			return nil, errors.New("Unspecific error on connection")
		} else if msg.Header.Type != dto.Type_TCP_CONNECTION_ESTABLISHED {
			transport.UnregisterChannel(instance.connectionId, instance.channel)
			return nil, errors.New(fmt.Sprintf("Unknown response type %v", msg.Header.Type))
		}

	case <-time.After(45 * time.Second):
		transport.UnregisterChannel(instance.connectionId, instance.channel)
		return nil, errors.New("Connection cannot be established within 30 seconds")
	}

	return instance, nil
}

func (this *ProxyConnection) Read(b []byte) (n int, err error) {
	if len(this.remainingBytes) > 0 {
		n := copy(b, this.remainingBytes)
		this.remainingBytes = this.remainingBytes[n:]
		return n, nil
	}

	msg, more := <-this.channel
	if !more {
		this.Close()
		return 0, io.EOF
	}

	switch msg.Header.Type {
	case dto.Type_TCP_CONNECTION_CLOSED:
		this.Close()
		return 0, io.EOF

	case dto.Type_INBOUND_DATA:
		data := msg.Payload.GetData()
		n := copy(b, data)
		if n < len(data) { // not all data is copied
			this.remainingBytes = data[n:]
		}
		return n, nil

	default:
		log.Println("Unknown type:", msg.Header.Type)
	}

	return 0, nil
}

func (this *ProxyConnection) Write(data []byte) (n int, err error) {
	if len(data) > 0 {
		payload := &dto.Payload{
			Data: data,
		}
		// send the message
		err := this.transport.Write(dto.Type_OUTBOUND_DATA, this.connectionId, payload)
		if err != nil {
			this.Close()
			return 0, err
		}
	}
	return len(data), nil
}

func (this *ProxyConnection) Close() error {
	this.transport.UnregisterChannel(this.connectionId, this.channel)
	return nil
}

func (c *ProxyConnection) LocalAddr() net.Addr {
	return nil
}

func (c *ProxyConnection) RemoteAddr() net.Addr {
	return nil
}

func (c *ProxyConnection) SetDeadline(t time.Time) error {
	return errors.New("Not implemented")
}

func (c *ProxyConnection) SetReadDeadline(t time.Time) error {
	return errors.New("Not implemented")
}

func (c *ProxyConnection) SetWriteDeadline(t time.Time) error {
	return errors.New("Not implemented")
}
