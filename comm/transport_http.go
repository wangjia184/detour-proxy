package comm

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"../dto"
)

type HttpTransport struct {
	uri             *url.URL
	running         bool // this flag tells goroutine if it should exits
	outboundReady   bool // this flag represents if outbound connection is established
	inboundReady    bool // this flag represents if inbound connection is established
	outboundChannel chan []byte
	client          *http.Client
	channels        map[int64]chan dto.Message
	mutex           sync.RWMutex
}

func NewHttpTransport(uri string) (Transport, error) {
	transport := new(HttpTransport)
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	transport.uri = u
	transport.running = true
	transport.outboundChannel = make(chan []byte)
	transport.channels = make(map[int64]chan dto.Message)
	transport.mutex = sync.RWMutex{}

	httpTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 300 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       300 * time.Second,
		DisableCompression:    false,
	}

	transport.client = &http.Client{
		Timeout:   12 * time.Hour,
		Transport: httpTransport,
	}

	go transport.inbound()
	go transport.outbound()
	return Transport(transport), nil
}

func (this *HttpTransport) Stop() {
	close(this.outboundChannel)
	this.running = false
}

func (this *HttpTransport) RegisterChannel(connectionID int64, channel chan dto.Message) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	originalChannel := this.channels[connectionID]
	this.channels[connectionID] = channel
	if originalChannel != nil {
		close(originalChannel)
	}
}

func (this *HttpTransport) UnregisterChannel(connectionID int64, channel chan dto.Message) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	originalChannel := this.channels[connectionID]
	if originalChannel == channel {
		delete(this.channels, connectionID)
		close(originalChannel)
	}
}

func (this *HttpTransport) getChannel(connectionID int64) chan dto.Message {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	return this.channels[connectionID]
}

func (this *HttpTransport) Write(msgType dto.Type, connectionID int64, payload *dto.Payload) error {
	if !this.running || !this.outboundReady {
		return errors.New("Proxy is unavailable")
	}

	bytes, err := dto.Encode(msgType, connectionID, payload)
	if err != nil {
		return err
	}

	this.outboundChannel <- bytes
	return nil
}

func (this *HttpTransport) outbound() {
	reader, writer := io.Pipe()
	defer reader.Close()

	request := &http.Request{
		Method:           "PUT",
		ProtoMajor:       1,
		ProtoMinor:       1,
		URL:              this.uri,
		TransferEncoding: []string{"chunked"},
		Body:             reader,
		ContentLength:    -1,
		Header:           make(http.Header),
	}
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")

	go func() {
		defer writer.Close()

		// header represents the length in little endian
		heartbeat := make([]byte, 1, 1)
		heartbeat[0] = 0

		idleTime := 0 * time.Second
		for this.running {

			select {
			case bytes, more := <-this.outboundChannel:
				{
					if more {

						// then send the chunk data
						_, err := writer.Write(bytes)
						if err != nil {
							if this.running {
								log.Println("The outbound goroutine exited because an error occurs :", err)
							}
							return
						}
						this.outboundReady = true
						idleTime = 50 * time.Millisecond

					} else {
						return // channel is closed
					}
				} // case end
			case <-time.After(idleTime):
				{
					if !this.running {
						return // normal exit
					}

					// send heartbeat if it idles
					_, err := writer.Write(heartbeat)
					if err != nil {
						log.Println("The outbound goroutine exited because an error occurs :", err)
						return
					}
					this.outboundReady = true
					idleTime = 20 * time.Second
				} // case end
			} // select
		} // for this.running
	}()

	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// client.Do returns after received the HTTP response.
	// Because we are keep sending HTTP request progressively
	// client.Do() does not return until an error occurs
	// Here for a successful request, it does not return till there is an error occurs
	this.outboundReady = false
	response, err := this.client.Do(request)
	this.outboundReady = false

	if err != nil {
		log.Println(err)
	} else {
		defer response.Body.Close()

		if response.StatusCode != 200 {
			log.Println("PUT", this.uri, "returned HTTP status code :", response.StatusCode)
		} else {
			body, _ := ioutil.ReadAll(response.Body)
			bodyString := string(body)
			log.Println(bodyString)
		}
	}

	if this.running {
		time.AfterFunc(time.Second, func() { this.outbound() })
	}
}

func (this *HttpTransport) inbound() {

	request := &http.Request{
		Method:     "GET",
		ProtoMajor: 1,
		ProtoMinor: 1,
		URL:        this.uri,
		Header:     make(http.Header),
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36")

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
				header, payload, err := dto.ReadMessageAndPayload(response.Body, length)
				if err != nil {
					log.Println(err)
					break
				}
				channel := this.getChannel(header.ConnectionID) // find the channel by connection id
				if channel == nil {
					channel = this.getChannel(0) // get the channel of connection id zero. which is defined as default
				}
				if channel != nil {
					msg := &dto.Message{
						Header:  header,
						Payload: payload,
					}
					channel <- *msg
				}
			}

			this.inboundReady = false
		}
	}

	if this.running {
		time.AfterFunc(time.Second, func() { this.inbound() })
	}
}
