package game

import (
	configModel "IsaacCoyote/common/config/model"
	"IsaacCoyote/common/isaac"
	"IsaacCoyote/pkg/coyote"
	"IsaacCoyote/pkg/coyote/enums"
	"container/list"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Game struct {
	config        *configModel.Game
	coyoteSession *coyote.Session

	isaacListener *isaac.GameListener
	playerInfo    playerInfo

	needContModeDecayCalc bool
	collStrengthAddA      int
	collStrengthAddB      int

	dequeLock  sync.Mutex
	pulseDeque *list.List
}

func (g *Game) Run() error {
	go g.dispatchPulse()
	go g.continuousMode()
	err := g.initCallbacks()
	if err != nil {
		return err
	}

	g.updateIndicator()
	return nil
}

func (g *Game) dispatchPulse() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if !g.coyoteSession.IsBound() {
			continue
		}

		g.dequeLock.Lock()
		segment := g.pulseDeque.Front()
		if segment == nil {
			g.dequeLock.Unlock()
			err := g.coyoteSession.SetStrength(enums.ChannelTypeA, enums.StrengthActionSetTo, 0)
			if err != nil {
				zap.L().Error("Failed to set strength A", zap.Error(err))
			}
			err = g.coyoteSession.SetStrength(enums.ChannelTypeB, enums.StrengthActionSetTo, 0)
			if err != nil {
				zap.L().Error("Failed to set strength B", zap.Error(err))
			}
			continue
		}
		g.pulseDeque.Remove(segment)
		g.dequeLock.Unlock()

		strengthA := segment.Value.(pulseSegment).StrengthA
		strengthB := segment.Value.(pulseSegment).StrengthB
		currStrengthData := g.coyoteSession.GetStrengthData()
		if strengthA > currStrengthData.MaxStrengthA {
			strengthA = currStrengthData.MaxStrengthA
		}
		if strengthB > currStrengthData.MaxStrengthB {
			strengthB = currStrengthData.MaxStrengthB
		}

		err := g.coyoteSession.SetStrength(enums.ChannelTypeA, enums.StrengthActionSetTo, strengthA)
		if err != nil {
			zap.L().Error("Failed to set strength A", zap.Error(err))
		}
		err = g.coyoteSession.SetStrength(enums.ChannelTypeB, enums.StrengthActionSetTo, strengthB)
		if err != nil {
			zap.L().Error("Failed to set strength B", zap.Error(err))
		}

		//g.coyoteSession.ClearPulse(enums.ChannelTypeA)
		//g.coyoteSession.ClearPulse(enums.ChannelTypeB)
		framesA := segment.Value.(pulseSegment).FramesA
		framesB := segment.Value.(pulseSegment).FramesB
		err = g.coyoteSession.AddPulse(enums.ChannelTypeA, framesA)
		if err != nil {
			zap.L().Error("failed to add pulse", zap.Error(err))
		}
		err = g.coyoteSession.AddPulse(enums.ChannelTypeB, framesB)
		if err != nil {
			zap.L().Error("failed to add pulse", zap.Error(err))
		}
	}
}

func (g *Game) initCallbacks() error {
	_ = g.isaacListener.RegisterCallback(isaac.GameStartEvent, func(callbackData interface{}) {
		startData := callbackData.(isaac.GameStartEventData)
		if !startData.IsContinue {
			g.reset()
		}
	})

	_ = g.isaacListener.RegisterCallback(isaac.GameEndEvent, func(callbackData interface{}) {
		g.reset()
	})

	_ = g.isaacListener.RegisterCallback(isaac.PlayerInfoUpdateEvent, func(callbackData interface{}) {
		data := callbackData.(isaac.PlayerInfoUpdateEventData)
		g.playerInfo.Health = data.Health
		g.playerInfo.MaxHealth = data.MaxHealth
		collectibles, err := parseCollectiblesString(data.Collectibles, g.isaacListener.ResourceManager)
		if err != nil {
			zap.L().Error("获取物品失败: ", zap.Error(err))
		}

		// Update collectibles and collectibles strength
		if g.playerInfo.collString != data.Collectibles {
			g.playerInfo.collString = data.Collectibles
			g.playerInfo.Collectibles = collectibles

			if g.config.OnNewCollectible.Enabled {
				g.collStrengthAddA = 0
				g.collStrengthAddB = 0
				for _, item := range collectibles {
					if quality := item.itemDetail.Quality; quality >= 0 {
						if config, ok := g.config.OnNewCollectible.StrengthConfig[quality]; ok {
							g.collStrengthAddA += config.StrengthAddA * item.num
							g.collStrengthAddB += config.StrengthAddB * item.num
						} else {
							zap.L().Info("未配置强度的物品: " + item.itemDetail.Name)
						}
					}
				}

				g.needContModeDecayCalc = true
			}
		}
	})

	_ = g.isaacListener.RegisterCallback(isaac.PlayerHurtEvent, func(interface{}) {
		if !g.config.OnHurt.Enabled || !g.coyoteSession.IsBound() {
			return
		}

		var pulseIndexA int
		var pulseIndexB int
		var strengthA int
		var strengthB int

		if g.config.OnHurt.StrengthOperator == configModel.INCREMENT {
			strengthA = g.getMinStrengthA() + g.config.OnHurt.StrengthA
			strengthB = g.getMinStrengthB() + g.config.OnHurt.StrengthB
		} else {
			strengthA = g.config.OnHurt.StrengthA
			strengthB = g.config.OnHurt.StrengthB
		}

		segmentList := list.New()

		for duration := g.config.OnHurt.Duration; duration >= 0; duration -= 200 {
			segment := pulseSegment{
				FramesA:   nextTwoPulseFrames(g.config.OnHurt.PulseA.PulseWaveform, &pulseIndexA),
				FramesB:   nextTwoPulseFrames(g.config.OnHurt.PulseB.PulseWaveform, &pulseIndexB),
				StrengthA: strengthA,
				StrengthB: strengthB,
			}
			segmentList.PushBack(segment)
		}
		g.dequeLock.Lock()
		g.pulseDeque = list.New() //clear
		g.pulseDeque.PushFrontList(segmentList)
		g.dequeLock.Unlock()
		g.needContModeDecayCalc = true
	})

	_ = g.isaacListener.RegisterCallback(isaac.PlayerDeathEvent, func(interface{}) {
		if !g.config.OnDeath.Enabled || !g.coyoteSession.IsBound() {
			return
		}

		var pulseIndexA int
		var pulseIndexB int
		var strengthA int
		var strengthB int

		if g.config.OnDeath.StrengthOperator == configModel.INCREMENT {
			strengthA = g.getMinStrengthA() + g.config.OnDeath.StrengthA
			strengthB = g.getMinStrengthB() + g.config.OnDeath.StrengthB
		} else {
			strengthA = g.config.OnDeath.StrengthA
			strengthB = g.config.OnDeath.StrengthB
		}

		segmentList := list.New()
		for duration := g.config.OnDeath.Duration; duration >= 0; duration -= 200 {
			segment := pulseSegment{
				FramesA:   nextTwoPulseFrames(g.config.OnDeath.PulseA.PulseWaveform, &pulseIndexA),
				FramesB:   nextTwoPulseFrames(g.config.OnDeath.PulseB.PulseWaveform, &pulseIndexB),
				StrengthA: strengthA,
				StrengthB: strengthB,
			}
			segmentList.PushBack(segment)
		}
		g.dequeLock.Lock()

		g.pulseDeque = list.New() //clear
		g.pulseDeque.PushFrontList(segmentList)
		g.dequeLock.Unlock()

		g.needContModeDecayCalc = true
		g.playerInfo = playerInfo{}
	})

	_ = g.isaacListener.RegisterCallback(isaac.ManualRestartEvent, func(callbackData interface{}) {
		if !g.config.OnManualRestart.Enabled || !g.coyoteSession.IsBound() {
			return
		}

		var pulseIndexA int
		var pulseIndexB int
		var strengthA int
		var strengthB int

		if g.config.OnManualRestart.StrengthOperator == configModel.INCREMENT {
			strengthA = g.config.BaseStrengthA + g.config.OnManualRestart.StrengthA
			strengthB = g.config.BaseStrengthB + g.config.OnManualRestart.StrengthB
		} else {
			strengthA = g.config.OnManualRestart.StrengthA
			strengthB = g.config.OnManualRestart.StrengthB
		}

		segmentList := list.New()
		for duration := g.config.OnManualRestart.Duration; duration >= 0; duration -= 200 {
			segment := pulseSegment{
				FramesA:   nextTwoPulseFrames(g.config.OnManualRestart.PulseA.PulseWaveform, &pulseIndexA),
				FramesB:   nextTwoPulseFrames(g.config.OnManualRestart.PulseB.PulseWaveform, &pulseIndexB),
				StrengthA: strengthA,
				StrengthB: strengthB,
			}
			segmentList.PushBack(segment)
		}
		g.dequeLock.Lock()
		g.pulseDeque = list.New() //clear
		g.pulseDeque.PushFrontList(segmentList)
		g.dequeLock.Unlock()

		g.needContModeDecayCalc = true
	})

	return nil
}

func (g *Game) continuousMode() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	var (
		lastDecayTime = time.Now()
		pulseIndexA   = 0
		pulseIndexB   = 0
		prevSegment   = pulseSegment{}
	)

	for range ticker.C {
		if !g.config.ContinuousMode.Enabled || !g.coyoteSession.IsBound() || g.pulseDeque.Len() >= 100 {
			continue
		}

		if g.needContModeDecayCalc {
			g.needContModeDecayCalc = false

			lastSegment := g.pulseDeque.Front()
			if lastSegment != nil {
				prevSegment = lastSegment.Value.(pulseSegment)
			} else {
				prevSegment = pulseSegment{
					StrengthA: g.coyoteSession.GetStrengthData().StrengthA,
					StrengthB: g.coyoteSession.GetStrengthData().StrengthB,
				}
			}

			pulseIndexB = 0
			pulseIndexA = 0
		}

		// set pulse frame
		segment := pulseSegment{
			FramesA: nextTwoPulseFrames(g.config.ContinuousMode.PulseA.PulseWaveform, &pulseIndexA),
			FramesB: nextTwoPulseFrames(g.config.ContinuousMode.PulseB.PulseWaveform, &pulseIndexB),
		}

		//set strength
		minA := g.getMinStrengthA()
		minB := g.getMinStrengthB()

		if g.config.ContinuousMode.DecayInterval > 0 {
			segment.StrengthA = prevSegment.StrengthA
			segment.StrengthB = prevSegment.StrengthB

			intervalDuration := time.Duration(g.config.ContinuousMode.DecayInterval) * time.Millisecond
			elapsed := time.Now().Sub(lastDecayTime)
			if elapsed >= intervalDuration {
				decayCount := int(elapsed / intervalDuration)
				segment.StrengthA -= g.config.ContinuousMode.DecayValue * decayCount
				segment.StrengthB -= g.config.ContinuousMode.DecayValue * decayCount

				lastDecayTime = lastDecayTime.Add(time.Duration(decayCount) * intervalDuration)
			}
		} else {
			segment.StrengthA, segment.StrengthB = minA, minB
		}

		segment.StrengthA = max(segment.StrengthA, minA)
		segment.StrengthB = max(segment.StrengthB, minB)

		g.dequeLock.Lock()
		g.pulseDeque.PushBack(segment)
		prevSegment = segment
		g.dequeLock.Unlock()
	}
}

func (g *Game) updateIndicator() {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		if !g.coyoteSession.IsBound() {
			continue
		}

		strengthData := g.coyoteSession.GetStrengthData()
		g.isaacListener.AddUpdateIndicatorMsg(strengthData.StrengthA, strengthData.StrengthB)
	}
}

func (g *Game) getMinStrengthA() int {
	return g.config.BaseStrengthA +
		g.config.StrengthPerHealthA*(g.playerInfo.MaxHealth-g.playerInfo.Health) +
		g.collStrengthAddA
}

func (g *Game) getMinStrengthB() int {
	return g.config.BaseStrengthB +
		g.config.StrengthPerHealthB*(g.playerInfo.MaxHealth-g.playerInfo.Health) +
		g.collStrengthAddB
}

func (g *Game) reset() {
	g.needContModeDecayCalc = true
	g.collStrengthAddA, g.collStrengthAddB = 0, 0
	g.playerInfo = playerInfo{}
}

func NewGame(config *configModel.Game, coyoteSession *coyote.Session, isaacListener *isaac.GameListener) *Game {
	return &Game{
		config:        config,
		coyoteSession: coyoteSession,
		isaacListener: isaacListener,

		pulseDeque: list.New(),
	}
}
