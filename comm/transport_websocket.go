package comm

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"../dto"

	"github.com/gorilla/websocket"
)

type WebSocketTransport struct {
	uri             *url.URL
	running         bool // this flag tells goroutine if it should exits
	outboundChannel chan []byte
	channels        map[int64]chan dto.Message
	mutex           sync.RWMutex
	lastDialError   string
}

func NewWebSocketTransport(uri string) (Transport, error) {
	this := new(WebSocketTransport)
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	this.uri = u
	this.running = true
	this.outboundChannel = make(chan []byte, 10)
	this.channels = make(map[int64]chan dto.Message)
	this.mutex = sync.RWMutex{}

	go this.inbound()
	//go transport.outbound()
	return Transport(this), nil
}

func (this *WebSocketTransport) Stop() {
	close(this.outboundChannel)
	this.running = false
}

func (this *WebSocketTransport) RegisterChannel(connectionID int64, channel chan dto.Message) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	originalChannel := this.channels[connectionID]
	this.channels[connectionID] = channel
	if originalChannel != nil {
		close(originalChannel)
	}
}

func (this *WebSocketTransport) UnregisterChannel(connectionID int64, channel chan dto.Message) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	originalChannel := this.channels[connectionID]
	if originalChannel == channel {
		delete(this.channels, connectionID)
		close(originalChannel)
	}
}

func (this *WebSocketTransport) getChannel(connectionID int64) chan dto.Message {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.channels[connectionID]
}

func (this *WebSocketTransport) Write(msgType dto.Type, connectionID int64, payload *dto.Payload) error {
	if !this.running {
		return errors.New("Proxy is unavailable")
	}

	bytes, err := dto.Encode(msgType, connectionID, payload)
	if err != nil {
		return err
	}

	this.outboundChannel <- bytes
	return nil
}

func (this *WebSocketTransport) inbound() {

	if !this.running {
		return
	}

	defer time.AfterFunc(50*time.Millisecond, func() {
		if this.running {
			this.inbound()
		}
	})

	httpHeader := make(http.Header)
	httpHeader.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")
	dialer := &websocket.Dialer{
		EnableCompression: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	}
	wsConn, _, err := dialer.Dial(this.uri.String(), httpHeader)
	if err != nil {
		if this.lastDialError != err.Error() { // avoid duplicate errors
			log.Println(err)
			this.lastDialError = err.Error()
		}
		return
	} else if len(this.lastDialError) != 0 {
		this.lastDialError = ""
		log.Println("Connected to", this.uri)
	}
	defer wsConn.Close()

	readerExitedChannel := make(chan bool)
	exited := false
	go (func() {
		defer (func() { exited = true })()

		for this.running && !exited {

			select {
			case bytes, more := <-this.outboundChannel:
				{
					if more {
						// then send the chunk data
						err := wsConn.WriteMessage(websocket.BinaryMessage, bytes)
						if err != nil {
							log.Println(err)
							return
						}
					} else {
						log.Println("Writer goroutine exited because outboundChannel was closed")
						return // channel is closed
					}
				} // case end
			case <-readerExitedChannel:
				{
					log.Println("Writer goroutine exited")
					return
				}
			case <-time.After(20 * time.Second):
				{
					// send an empty text message to keep connection alive
					err := wsConn.WriteMessage(websocket.TextMessage, nil)
					if err != nil {
						log.Println(err)
						return
					}
				} // case end
			} // select

		} //for {
	})()

	for this.running && !exited {
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
			msg, err := dto.Decode(buffer)
			if err != nil {
				log.Println(err)
				break
			}
			channel := this.getChannel(msg.Header.ConnectionID) // find the channel by connection id
			if channel == nil {
				channel = this.getChannel(0) // get the channel of connection id zero. which is defined as default
			}
			if channel != nil {
				channel <- *msg
			}
		}
	}

	exited = true
	close(readerExitedChannel)
	log.Println("Reader goroutine exited")
	/*
		this.inboundReady = false
		response, err := this.client.Do(request)
		if err != nil {
			log.Println("client.Do", err)
		} else {
			defer response.Body.Close()

			if response.StatusCode != 200 {
				log.Println("GET", this.uri, "returned HTTP status code :", response.StatusCode)
			} else {
				lengthByte := make([]byte, 1, 1)

				for {
					cnt, err := io.ReadFull(response.Body, lengthByte)
					if err != nil && cnt != 1 {
						log.Println("response.Body.Read", err)
						break
					}

					this.inboundReady = true
					if lengthByte[0] == 0 { // heartbeat
						continue
					}

					length := int(lengthByte[0])

				}

				this.inboundReady = false
			}
		}

		if this.running {
			time.AfterFunc(time.Second, func() { this.inbound() })
		}
	*/
}
