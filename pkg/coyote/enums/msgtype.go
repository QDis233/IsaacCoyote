package enums

import (
	"encoding/json"
	"fmt"
)

type MsgType string

const (
	MsgTypeHeartBeat MsgType = "heartbeat"
	MsgTypeBind      MsgType = "bind"
	MsgTypeMessage   MsgType = "msg"
	MsgTypeBreak     MsgType = "break"
	MsgTypeError     MsgType = "error"

	MsgTypeUnknown MsgType = "unknown"
)

func GetMsgType(rawMsgType string) MsgType {
	switch rawMsgType {
	case "heartbeat":
		return MsgTypeHeartBeat
	case "bind":
		return MsgTypeBind
	case "msg":
		return MsgTypeMessage
	case "break":
		return MsgTypeBreak
	case "error":
		return MsgTypeError
	}
	return MsgTypeUnknown
}

func (m *MsgType) String() string {
	return string(*m)
}

func (m *MsgType) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m *MsgType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch s {
	case "heartbeat":
		*m = MsgTypeHeartBeat
	case "bind":
		*m = MsgTypeBind
	case "msg":
		*m = MsgTypeMessage
	case "break":
		*m = MsgTypeBreak
	case "error":
		*m = MsgTypeError
	default:
		return fmt.Errorf("invalid MsgType: %s", s)
	}
	return nil
}
