package coyote

import (
	"fmt"
	"github.com/olahol/melody"
	"go.uber.org/zap"
	"net/http"
)

type MsgHandler func(s *melody.Session, msg []byte)
type DisconnectHandler func(s *melody.Session)
type ConnectHandler func(s *melody.Session)

type Server struct {
	IsRunning bool

	melody *melody.Melody

	config            *Config
	msgHandler        MsgHandler
	connHandler       ConnectHandler
	disconnectHandler DisconnectHandler
}

func (s *Server) Run() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := s.melody.HandleRequest(w, r)
		if err != nil {
			zap.L().Error("Failed to handle request", zap.Error(err))
		}
	})
	s.melody.HandleMessage(s.msgHandler)
	s.melody.HandleConnect(s.connHandler)
	s.melody.HandleDisconnect(s.disconnectHandler)

	s.IsRunning = true
	err := http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), nil)
	if err != nil {
		s.IsRunning = false
		return err
	}
	return nil
}

func NewCoyoteServer(config *Config, connHandler ConnectHandler, disconnectHandler DisconnectHandler, msgHandler MsgHandler) *Server {
	return &Server{
		config: config,

		melody:            melody.New(),
		msgHandler:        msgHandler,
		connHandler:       connHandler,
		disconnectHandler: disconnectHandler,
	}
}
