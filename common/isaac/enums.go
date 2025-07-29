package isaac

type Event string

const (
	ModInitEvent          Event = "ModInitEvent"
	PlayerHurtEvent       Event = "PlayerHurtEvent"
	PlayerDeathEvent      Event = "PlayerDeathEvent"
	ManualRestartEvent    Event = "ManualRestartEvent"
	GameStartEvent        Event = "GameStartEvent"
	GameExitEvent         Event = "GameExitEvent"
	GameEndEvent          Event = "GameEndEvent"
	NewCollectibleEvent   Event = "NewCollectibleEvent"
	PlayerInfoUpdateEvent Event = "PlayerInfoUpdateEvent"
)

func (e Event) String() string {
	return string(e)
}

type MsgType string

const (
	EventMsg           = "event"
	UpdateIndicatorMsg = "update_indicator"
	ConnectMsg         = "connect"
	HeartbeatMsg       = "heartbeat"
)
