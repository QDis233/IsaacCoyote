package isaac

import (
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"time"
)

type CallbackFunc func(callbackData interface{})

type GameListener struct {
	modDataPath string

	ResourceManager *ResourceManager
	callbacks       map[Event][]CallbackFunc

	currModData          ModData
	lastRecHeartbeatTime time.Time
	IsConnected          bool
	msgBuffer            []ModMessage
}

func (g *GameListener) Run() error {
	err := g.ResourceManager.LoadResources()
	if err != nil {
		zap.L().Error("请检查资源文件", zap.Error(err))
		return err
	}

	pid := waitForProcess("isaac-ng.exe")
	modDataPath, err := getModDataFile(pid)
	if err != nil {
		zap.L().Error("获取数据文件失败", zap.Error(err))
		return err
	}
	g.modDataPath = modDataPath

	g.AddMessage(ModMessage{
		Type: ConnectMsg,
	})
	err = g.Write()
	if err != nil {
		zap.L().Error("写入数据文件失败", zap.Error(err))
		return err
	}
	g.IsConnected = true
	g.lastRecHeartbeatTime = time.Now()
	g.triggerCallback(ModInitEvent, nil)

	var lastHeartbeatTime time.Time
	ticker := time.NewTicker(256 * time.Millisecond)
	defer ticker.Stop()

	for g.IsConnected {
		<-ticker.C
		err = g.FetchData()
		if err != nil {
			zap.L().Error("读取数据失败", zap.Error(err))
			continue
		}
		g.statistics()

		if time.Since(lastHeartbeatTime) > 2*time.Second {
			g.addHeartbeatMsg()
			lastHeartbeatTime = time.Now()
		}

		if len(g.msgBuffer) != 0 {
			err = g.Write()
			if err != nil {
				continue
			}
		}
		g.checkConnection()
	}
	return nil
}

func (g *GameListener) statistics() {
	var eventList []EventMessageData
	var deathFlag bool
	for _, message := range g.currModData.Send {
		switch message.Type {
		case EventMsg:
			eventData := message.Message.(EventMessageData)
			if eventData.Type == PlayerDeathEvent.String() {
				deathFlag = true
			}
			eventList = append(eventList, message.Message.(EventMessageData))
		case HeartbeatMsg:
			g.lastRecHeartbeatTime = time.Now()
		default:
			return
		}
	}
	g.dispatchEvent(eventList, deathFlag)
}

func (g *GameListener) checkConnection() {
	if time.Since(g.lastRecHeartbeatTime) > 2*time.Second {
		g.IsConnected = false
	} else {
		g.IsConnected = true
	}
}

func (g *GameListener) Write() error {
	data := ModData{
		Send:    make([]ModMessage, 0),
		Receive: g.msgBuffer,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	g.msgBuffer = make([]ModMessage, 0)
	return os.WriteFile(g.modDataPath, jsonData, 0644)
}

func (g *GameListener) AddMessage(message ModMessage) {
	g.msgBuffer = append(g.msgBuffer, message)
}

func (g *GameListener) addHeartbeatMsg() {
	g.AddMessage(ModMessage{
		Type: HeartbeatMsg,
	})
}

func (g *GameListener) AddUpdateIndicatorMsg(strengthA int, strengthB int) {
	g.AddMessage(ModMessage{
		Type: UpdateIndicatorMsg,
		Message: UpdateIndicatorData{
			StrengthA: strengthA,
			StrengthB: strengthB,
		},
	})
}

func (g *GameListener) FetchData() error {
	rawData, err := os.ReadFile(g.modDataPath)
	if err != nil {
		return err
	}

	var data ModData
	_ = json.Unmarshal(rawData, &data)
	g.currModData = data
	return nil
}

func (g *GameListener) dispatchEvent(eventList []EventMessageData, deathFlag bool) {
	for _, eventData := range eventList {
		switch eventData.Type {
		case PlayerHurtEvent.String():
			if !deathFlag {
				g.triggerCallback(PlayerHurtEvent, eventData.Data.(PlayerHurtEventData))
			}
			break
		case PlayerInfoUpdateEvent.String():
			g.triggerCallback(PlayerInfoUpdateEvent, eventData.Data.(PlayerInfoUpdateEventData))
			break
		case PlayerDeathEvent.String():
			g.triggerCallback(PlayerDeathEvent, nil)
			break
		case ManualRestartEvent.String():
			g.triggerCallback(ManualRestartEvent, nil)
			break
		case GameStartEvent.String():
			g.triggerCallback(GameStartEvent, eventData.Data.(GameStartEventData))
			break
		case GameExitEvent.String():
			g.triggerCallback(GameExitEvent, nil)
			break
		case GameEndEvent.String():
			g.triggerCallback(GameEndEvent, nil)
			break
		}
	}
}

func (g *GameListener) triggerCallback(eventType Event, callbackData interface{}) {
	if g.callbacks == nil {
		return
	}
	callbacks, ok := g.callbacks[eventType]
	if !ok || len(callbacks) == 0 {
		return
	}
	for _, callback := range callbacks {
		go callback(callbackData)
	}
	return
}

func (g *GameListener) RegisterCallback(eventType Event, callback CallbackFunc) error {
	if g.callbacks == nil {
		g.callbacks = make(map[Event][]CallbackFunc)
	}
	g.callbacks[eventType] = append(g.callbacks[eventType], callback)
	return nil
}

func NewGameListener() *GameListener {
	return &GameListener{
		ResourceManager: NewResourceManager(),
	}
}
