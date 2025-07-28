package coyote

import (
	"IsaacCoyote/pkg/coyote/enums"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/olahol/melody"
	"go.uber.org/zap"
)

type Coyote struct {
	config *Config

	wsServer  *Server
	callbacks map[enums.ServerEvent][]func(callbackData CallbackData[any])
	sessions  map[string]*Session //map[clientID]*Session
}

func (c *Coyote) IsRunning() bool {
	if c.wsServer == nil {
		return false
	}
	return c.wsServer.IsRunning
}

func (c *Coyote) Run() error {
	if c.wsServer == nil {
		c.wsServer = NewCoyoteServer(c.config, c.connectHandler, c.disconnectHandler, c.msgHandler)
	}
	if c.wsServer.IsRunning {
		return AlreadyRunningError{
			Message: "  already running",
		}
	}
	return c.wsServer.Run()
}

func (c *Coyote) GetSessionByClientID(clientID string) (*Session, error) {
	session := c.sessions[clientID]
	if session == nil {
		return nil, SessionNotFoundError{
			Message: fmt.Sprintf("Session not found for clientID: %s", clientID),
		}
	}
	return session, nil
}

func (c *Coyote) RegisterCallback(eventType enums.ServerEvent, callback func(callbackData CallbackData[any])) {
	c.callbacks[eventType] = append(c.callbacks[eventType], callback)
}

func (c *Coyote) NewSession() *Session {
	clientID := uuid.New().String()
	session := NewCoyoteSession(clientID, c.config)
	c.sessions[clientID] = session
	return session
}

func (c *Coyote) sendMsg(s *melody.Session, msg WSMessage) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		zap.L().Error("failed to marshal error message", zap.Error(err))
		return
	}
	err = s.Write(jsonMsg)
	if err != nil {
		zap.L().Error("failed to close session", zap.Error(err))
		return
	}
}

func (c *Coyote) connectHandler(s *melody.Session) {
	clientID := s.Request.RequestURI[1:]
	session := c.sessions[clientID]
	if session == nil {
		errMsg := WSMessage{
			Type:     enums.MsgTypeError,
			ClientID: clientID,
			MsgData:  enums.RetCodeReceiverOffline.String(),
		}
		c.sendMsg(s, errMsg)
		_ = s.Close()
		return
	}

	if session.IsBound() {
		errMsg := WSMessage{
			Type:     enums.MsgTypeError,
			ClientID: clientID,
			MsgData:  enums.RetCodeClientIDAlreadyUsed.String(),
		}
		c.sendMsg(s, errMsg)
		_ = s.Close()
	}

	c.dispatchEvent(enums.OnConnect, clientID)
	session.SetWSSession(s)
	bindMsg := WSMessage{
		Type:     enums.MsgTypeBind,
		ClientID: clientID,
		MsgData:  enums.MsgHeadTargetID.String(),
	}
	err := session.SendMessage(bindMsg)
	if err != nil {
		zap.L().Error("failed to bind", zap.Error(err))
	}
}

func (c *Coyote) disconnectHandler(s *melody.Session) {
	clientId, exists := s.Get("clientID")
	if !exists {
		return
	}
	c.sessions[clientId.(string)].Disconnect()
	c.dispatchEvent(enums.OnDisconnect, clientId.(string))
}

func (c *Coyote) msgHandler(s *melody.Session, rawMsg []byte) {
	message := WSMessage{}
	err := json.Unmarshal(rawMsg, &message)
	if err != nil {
		msg := WSMessage{
			MsgData: enums.RetCodeInvalidMessageFormat.String(),
			Type:    enums.MsgTypeError,
		}
		jsonMsg, _ := msg.ToJSON()
		_ = s.Write(jsonMsg)
		return
	}

	session := c.sessions[message.ClientID]
	if session == nil {
		zap.L().Error("Session not found", zap.String("clientID", message.ClientID))
		return
	}
	c.dispatchEvent(enums.OnMessageReceived, session.clientID)

	switch message.Type {
	case enums.MsgTypeBind:
		err = session.handleBind(message)
		if err != nil {
			zap.L().Error("Failed to bind", zap.Error(err))
		}
	case enums.MsgTypeMessage:
		err = session.handleMsg(message)
		if err != nil {
			return
		}
	case enums.MsgTypeHeartBeat:
		session.handleHeartBeat(message)
	case enums.MsgTypeBreak:
		session.handleBreak(message)
	case enums.MsgTypeError:
		zap.L().Error("Received error message", zap.String("message", message.MsgData))
	default:
		errMsg := WSMessage{
			MsgData: enums.RetCodeInternalError.String(),
			Type:    enums.MsgTypeError,
		}
		c.sendMsg(s, errMsg)
	}
}

func (c *Coyote) dispatchEvent(eventType enums.ServerEvent, clientID string) {
	callbacks := c.callbacks[eventType]
	for _, callback := range callbacks {
		callbackData := CallbackData[any]{
			CallbackData: clientID,
		}
		go callback(callbackData)
	}
}

func NewCoyote(config *Config) *Coyote {
	return &Coyote{
		wsServer:  nil,
		config:    config,
		callbacks: make(map[enums.ServerEvent][]func(callbackData CallbackData[any])),
		sessions:  make(map[string]*Session),
	}
}
