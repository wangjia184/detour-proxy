package comm

import (
	"../dto"
)

type Transport interface {
	Write(msgType dto.Type, connectionID int64, payload *dto.Payload) error
	RegisterChannel(connectionID int64, channel chan dto.Message)
	UnregisterChannel(connectionID int64, channel chan dto.Message)
}
