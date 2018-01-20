package dto

import (
	"errors"
	"io"

	"github.com/golang/protobuf/proto"
)

type Message struct {
	Header  *MessageHeader
	Payload *Payload
}

func Encode(msgType Type, connectionID int64, payload *Payload) ([]byte, error) {

	var payloadBytes []byte
	var err error
	if payload != nil {
		payloadBytes, err = proto.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	// TODO : encoding the bytes

	header := &MessageHeader{
		Type:         msgType,
		ConnectionID: connectionID,
		Mode:         Mode_NONE,
		Length:       int32(len(payloadBytes)),
	}
	headerBytes, err := proto.Marshal(header)
	if err != nil {
		return nil, err
	}

	if len(headerBytes) > 255 {
		return nil, errors.New("Header length should never be greater than 255")
	}

	// header represents the length in little endian
	bytes := make([]byte, 1, len(headerBytes)+len(payloadBytes)+1)
	bytes[0] = uint8(len(headerBytes))
	bytes = append(bytes, headerBytes...)
	bytes = append(bytes, payloadBytes...)

	return bytes, nil
}

func DecodeHeader(b []byte) (*MessageHeader, error) {

	if len(b) < 1 {
		return nil, nil
	}

	headerLength := int(b[0])
	if len(b) < 1+headerLength {
		return nil, errors.New("Insufficient buffer to decode")
	}

	slice := b[1 : headerLength+1]
	header := &MessageHeader{}
	err := proto.Unmarshal(slice, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func Decode(b []byte) (*Message, error) {

	msg := &Message{}
	if len(b) < 1 {
		return nil, nil
	}

	headerLength := int(b[0])
	if len(b) < 1+headerLength {
		return nil, errors.New("Insufficient buffer to decode")
	}

	slice := b[1 : headerLength+1]
	msg.Header = &MessageHeader{}
	err := proto.Unmarshal(slice, msg.Header)
	if err != nil {
		return nil, err
	}

	slice = b[headerLength+1:]
	msg.Payload = &Payload{}
	err = proto.Unmarshal(slice, msg.Payload)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func ReadMessageAndPayload(reader io.Reader, headerLength int) (*MessageHeader, *Payload, error) {
	headerBytes := make([]byte, headerLength, headerLength)
	n, err := io.ReadFull(reader, headerBytes)
	if err != nil {
		return nil, nil, err
	}
	if n != headerLength {
		return nil, nil, errors.New("Not enough data read")
	}

	header := &MessageHeader{}
	err = proto.Unmarshal(headerBytes, header)
	if err != nil {
		return nil, nil, err
	}

	if header.Length > 1024*1024*10 {
		return nil, nil, errors.New("Payload size is too large")
	}
	payloadBytes := make([]byte, header.Length, header.Length)
	n, err = io.ReadFull(reader, payloadBytes)
	if err != nil {
		return nil, nil, err
	}
	if n != int(header.Length) {
		return nil, nil, errors.New("Not enough data read")
	}

	// TODO : decoding the payload

	payload := &Payload{}
	err = proto.Unmarshal(payloadBytes, payload)
	if err != nil {
		return nil, nil, err
	}

	return header, payload, nil
}
