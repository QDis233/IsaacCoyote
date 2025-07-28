package isaac

import "encoding/json"

type ModData struct {
	Send    []ModMessage `json:"send"`
	Receive []ModMessage `json:"receive"`
}

type ModMessage struct {
	Type       string      `json:"type"`
	Message    interface{} `json:"message"`
	FrameCount int64       `json:"frameCount"`
}

func (m *ModMessage) UnmarshalJSON(data []byte) error {
	type Alias ModMessage
	aux := &struct {
		Message json.RawMessage `json:"message"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch m.Type {
	case EventMsg:
		var eventMsgData EventMessageData
		if err := json.Unmarshal(aux.Message, &eventMsgData); err != nil {
			return err
		}
		m.Message = eventMsgData
	case HeartbeatMsg:
		m.Message = nil
	default:
		m.Message = nil
	}
	return nil
}

type EventMessageData struct {
	Type string
	Data interface{}
}

func (e *EventMessageData) UnmarshalJSON(data []byte) error {
	var eventMsgData struct {
		Type string
		Data json.RawMessage
	}
	if err := json.Unmarshal(data, &eventMsgData); err != nil {
		return err
	}
	e.Type = eventMsgData.Type

	switch eventMsgData.Type {
	case "PlayerHurtEvent":
		var playerHurtEventData PlayerHurtEventData
		if err := json.Unmarshal(eventMsgData.Data, &playerHurtEventData); err != nil {
			return err
		}
		e.Data = playerHurtEventData
	case "NewCollectibleEvent":
		var newCollectibleEventData NewCollectibleEventData
		if err := json.Unmarshal(eventMsgData.Data, &newCollectibleEventData); err != nil {
			return err
		}
		e.Data = newCollectibleEventData
	case "PlayerInfoUpdateEvent":
		var playerInfoUpdateEventData PlayerInfoUpdateEventData
		if err := json.Unmarshal(eventMsgData.Data, &playerInfoUpdateEventData); err != nil {
			return err
		}
		e.Data = playerInfoUpdateEventData
	default:
		e.Data = nil
	}
	return nil
}

type PlayerHurtEventData struct {
	PlayerName string  `json:"playerName"`
	Damage     float64 `json:"damage"`
	Flags      int     `json:"flags"`
	Source     int     `json:"source"`
}

type NewCollectibleEventData struct {
	Name    string `json:"name"`
	ID      int    `json:"id"`
	Quality int    `json:"quality"`
}

type PlayerInfoUpdateEventData struct {
	Health    int `json:"health"`
	MaxHealth int `json:"maxHealth"`
}

type UpdateIndicatorData struct {
	StrengthA int `json:"strengthA"`
	StrengthB int `json:"strengthB"`
}
