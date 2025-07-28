package enums

import (
	"encoding/json"
	"fmt"
)

type MsgType int

const (
	MsgTypeHeartBeat MsgType = iota
	MsgTypeBind
	MsgTypeMessage
	MsgTypeBreak
	MsgTypeError
	MsgTypeUnknown
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
	switch *m {
	case MsgTypeHeartBeat:
		return "heartbeat"
	case MsgTypeBind:
		return "bind"
	case MsgTypeMessage:
		return "msg"
	case MsgTypeBreak:
		return "break"
	case MsgTypeError:
		return "error"
	case MsgTypeUnknown:
		return ""
	}
	return ""
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
