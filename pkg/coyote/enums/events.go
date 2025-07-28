package enums

type ServerEvent int

const (
	OnConnect ServerEvent = iota
	OnDisconnect
	OnMessageReceived
)

type SessionEvent int

const (
	OnSessionHeartBeat SessionEvent = iota
	OnSessionBind
	OnSessionMessageReceived
	OnSessionFeedback
	OnSessionStrengthChange
	OnSessionBreak
	OnSessionError
)
