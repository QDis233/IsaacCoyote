package enums

import "encoding/json"

type MsgHead int

const (
	MsgHeadTargetID MsgHead = iota
	MsgHeadDGLab
	MsgHeadStrength
	MsgHeadPulse
	MsgHeadClear
	MsgHeadFeedback
	MsgHeadUnknown
)

func (m MsgHead) String() string {
	switch m {
	case MsgHeadTargetID:
		return "targetId"
	case MsgHeadDGLab:
		return "DGLAB"
	case MsgHeadStrength:
		return "strength"
	case MsgHeadPulse:
		return "pulse"
	case MsgHeadClear:
		return "clear"
	case MsgHeadFeedback:
		return "feedback"
	case MsgHeadUnknown:
		return "unknown"
	}
	return ""
}

func (m MsgHead) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func GetMsgHead(s string) MsgHead {
	switch s {
	case "targetId":
		return MsgHeadTargetID
	case "DGLAB":
		return MsgHeadDGLab
	case "strength":
		return MsgHeadStrength
	case "pulse":
		return MsgHeadPulse
	case "clear":
		return MsgHeadClear
	case "feedback":
		return MsgHeadFeedback
	}
	return MsgHeadUnknown
}
