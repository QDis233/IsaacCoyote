package enums

import "encoding/json"

type MsgHead string

const (
	MsgHeadTargetID MsgHead = "targetId"
	MsgHeadDGLab    MsgHead = "DGLAB"
	MsgHeadStrength MsgHead = "strength"
	MsgHeadPulse    MsgHead = "pulse"
	MsgHeadClear    MsgHead = "clear"
	MsgHeadFeedback MsgHead = "feedback"
	MsgHeadUnknown  MsgHead = ""
)

func (m MsgHead) String() string {
	return string(m)
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
