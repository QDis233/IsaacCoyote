package coyote

import (
	"IsaacCoyote/pkg/coyote/enums"
	"encoding/json"
	"fmt"
	"github.com/olahol/melody"
	"go.uber.org/zap"
	"strings"
	"time"
)

type Session struct {
	wsSession *melody.Session
	config    *Config
	clientID  string
	targetID  string

	strengthData StrengthData

	lastHeartbeatTime time.Time
	callbacks         map[enums.SessionEvent][]func(s *Session, callbackData CallbackData[any])
	isBound           bool
}

func (s *Session) handleBind(message WSMessage) error {
	bindMsg := WSMessage{
		ClientID: message.ClientID,
		MsgData:  enums.RetCodeSuccess.String(),
		TargetID: message.TargetID,
		Type:     enums.MsgTypeBind,
	}

	if s.IsBound() {
		bindMsg.MsgData = enums.RetCodeClientIDAlreadyUsed.String()
		return s.SendMessage(bindMsg)
	}

	s.targetID = message.TargetID
	err := s.SendMessage(bindMsg)
	if err != nil {
		zap.L().Error("Failed to Bind", zap.Error(err))
		return err
	}

	s.isBound = true
	s.dispatchEvent(enums.OnSessionBind, message, s.targetID)
	return nil
}

func (s *Session) handleHeartBeat(message WSMessage) {
	s.lastHeartbeatTime = time.Now()
	s.dispatchEvent(enums.OnSessionHeartBeat, message, nil)
}

func (s *Session) handleBreak(message WSMessage) {
	s.targetID = ""
	s.dispatchEvent(enums.OnSessionBreak, message, nil)
}

func (s *Session) handleError(message WSMessage) {
	s.dispatchEvent(enums.OnSessionError, message, nil)
}

func (s *Session) handleMsg(message WSMessage) error {
	msgHead := enums.GetMsgHead(strings.Split(message.MsgData, "-")[0])

	s.dispatchEvent(enums.OnSessionMessageReceived, message, msgHead)

	switch msgHead {
	case enums.MsgHeadStrength:
		return s.updateStrength(message)
	case enums.MsgHeadFeedback:
		return s.handleFeedback(message)
	default:
		return nil // ignore other type
	}
}

func (s *Session) updateStrength(message WSMessage) error {
	result, err := ParseStrengthData(message.MsgData)
	if err != nil {
		return err
	}
	s.strengthData = StrengthData{
		StrengthA:    result[0],
		StrengthB:    result[1],
		MaxStrengthA: result[2],
		MaxStrengthB: result[3],
	}

	s.dispatchEvent(enums.OnSessionStrengthChange, message, s.strengthData)
	return nil
}

func (s *Session) handleFeedback(message WSMessage) error {
	buttonIndex, err := ParseFeedbackData(message.MsgData)
	if err != nil {
		return err
	}
	s.dispatchEvent(enums.OnSessionFeedback, message, buttonIndex)
	return nil
}

func (s *Session) SendMessage(message WSMessage) error {
	if s.wsSession == nil {
		return NoWSSessionError{Message: "No Websocket connection bound to this session"}
	}

	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if len(jsonMsg) > 1950 {
		return TooLongMessageError{
			Message: "message length larger than 1950",
		}
	}
	return s.wsSession.Write(jsonMsg)
}

func (s *Session) IsBound() bool {
	if s.wsSession == nil {
		return false
	}
	// 未实装的 heartbeat ?
	//if time.Since(s.lastHeartbeatTime) > 2*time.Second {
	//	return false
	//}
	return !s.wsSession.IsClosed() && s.isBound
}

func (s *Session) Disconnect() {
	if s.wsSession != nil && !s.wsSession.IsClosed() {
		_ = s.SendMessage(WSMessage{
			ClientID: s.clientID,
			Type:     enums.MsgTypeBreak,
			TargetID: s.clientID,
			MsgData:  "",
		})
		_ = s.wsSession.Close()
	}
	s.isBound = false
	s.targetID = ""
}

func (s *Session) SetWSSession(wsSession *melody.Session) {
	s.wsSession = wsSession
	s.wsSession.Set("clientID", s.clientID)
}

func (s *Session) RegisterCallback(eventType enums.SessionEvent, callback func(session *Session, callbackData CallbackData[any])) {
	s.callbacks[eventType] = append(s.callbacks[eventType], callback)
}

func (s *Session) dispatchEvent(eventType enums.SessionEvent, message WSMessage, data any) {
	callbacks := s.callbacks[eventType]
	for _, callback := range callbacks {
		callbackData := CallbackData[any]{
			Message:      &message,
			CallbackData: data,
		}
		go callback(s, callbackData)
	}
}

func (s *Session) GetQRCodeContent() string {
	uri := fmt.Sprintf("ws://%s:%d", s.config.Address, s.config.Port)
	return fmt.Sprintf("https://www.dungeon-lab.com/app-download.php#DGLAB-SOCKET#%s/%s", uri, s.clientID)
}

func (s *Session) WaitForBind() {
	for !s.IsBound() {
		time.Sleep(time.Millisecond * 1000)
	}
}

func (s *Session) SetStrength(channel enums.ChannelType, action enums.StrengthAction, strength int) error {
	if !s.IsBound() {
		return NotBindError{
			Message: "Session is not bound",
		}
	}

	msg := WSMessage{
		ClientID: s.clientID,
		TargetID: s.targetID,
		MsgData:  fmt.Sprintf("strength-%d+%d+%d", channel, action, strength),
		Type:     enums.MsgTypeMessage,
	}

	return s.SendMessage(msg)
}

func (s *Session) ClearPulse(channel enums.ChannelType) error {
	if !s.IsBound() {
		return NotBindError{
			Message: "Session is not bound",
		}
	}
	msg := WSMessage{
		ClientID: s.clientID,
		TargetID: s.targetID,
		MsgData:  fmt.Sprintf("clear-%d", channel),
		Type:     enums.MsgTypeMessage,
	}

	return s.SendMessage(msg)
}

func (s *Session) AddPulse(channel enums.ChannelType, waveform PulseWaveform) error {
	if !s.IsBound() {
		return NotBindError{
			Message: "Session is not bound",
		}
	}
	if waveform == nil || len(waveform) == 0 {
		return nil
	}

	rawPulseData, err := waveform.Marshal()
	if err != nil {
		return err
	}
	pulseData, err := json.Marshal(rawPulseData)
	if err != nil {
		return err
	}

	msg := WSMessage{
		ClientID: s.clientID,
		TargetID: s.targetID,
		MsgData:  fmt.Sprintf("pulse-%s:%s", channel.String(), strings.ReplaceAll(string(pulseData), `\`, "")),
		Type:     enums.MsgTypeMessage,
	}
	return s.SendMessage(msg)
}

func (s *Session) AddPulseFrame(channel enums.ChannelType, frames PulseFrame) error {
	if !s.IsBound() {
		return NotBindError{
			Message: "Session is not bound",
		}
	}
	rawPulseData, err := frames.Marshal()
	if err != nil {
		return err
	}
	pulseData, err := json.Marshal([]string{rawPulseData})
	if err != nil {
		return err
	}

	msg := WSMessage{
		ClientID: s.clientID,
		TargetID: s.targetID,
		MsgData:  fmt.Sprintf("pulse-%s:%s", channel.String(), strings.ReplaceAll(string(pulseData), `\`, "")),
		Type:     enums.MsgTypeMessage,
	}
	return s.SendMessage(msg)
}

func (s *Session) GetStrengthData() StrengthData {
	return s.strengthData
}

func NewCoyoteSession(clientID string, config *Config) *Session {
	return &Session{
		clientID: clientID,
		config:   config,

		callbacks: make(map[enums.SessionEvent][]func(s *Session, callbackData CallbackData[any])),
	}
}
