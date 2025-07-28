package game

import (
	configModel "IsaacCoyote/common/config/model"
	isaac2 "IsaacCoyote/common/isaac"
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

	isaacListener *isaac2.GameListener
	playerInfo    playerInfo

	needContModeDecayCalc bool
	collStrengthAddA      int
	collStrengthAddB      int

	dequeLock  sync.Mutex
	pulseDeque *list.List
}

func (g *Game) Run() error {
	go g.dispatchPulse()
	if g.config.ContinuousMode.Enabled {
		go g.continuousMode()
	}
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
	//_ = g.isaacListener.RegisterCallback(isaac.GameStartEvent, func(callbackData interface{}) {
	//	g.needContModeDecayCalc = true
	//	g.playerInfo = playerInfo{}
	//})

	_ = g.isaacListener.RegisterCallback(isaac2.NewCollectibleEvent, func(callbackData interface{}) {
		collData := callbackData.(isaac2.NewCollectibleEventData)
		if g.config.ContinuousMode.OnNewCollectible.Enabled {
			g.collStrengthAddA += g.config.ContinuousMode.OnNewCollectible.StrengthConfig[collData.Quality].StrengthA
			g.collStrengthAddB += g.config.ContinuousMode.OnNewCollectible.StrengthConfig[collData.Quality].StrengthB
		}
	})

	_ = g.isaacListener.RegisterCallback(isaac2.GameExitEvent, func(callbackData interface{}) {
		g.needContModeDecayCalc = true
		g.collStrengthAddA, g.collStrengthAddB = 0, 0
		g.playerInfo = playerInfo{}
	})

	_ = g.isaacListener.RegisterCallback(isaac2.PlayerInfoUpdateEvent, func(callbackData interface{}) {
		data := callbackData.(isaac2.PlayerInfoUpdateEventData)
		g.playerInfo.Health = data.Health
		g.playerInfo.MaxHealth = data.MaxHealth
	})

	_ = g.isaacListener.RegisterCallback(isaac2.PlayerHurtEvent, func(interface{}) {
		if !g.config.OnHurtMode.Enabled || !g.coyoteSession.IsBound() {
			return
		}

		var pulseIndexA int
		var pulseIndexB int

		segmentList := list.New()
		for duration := g.config.OnHurtMode.Duration; duration >= 0; duration -= 200 {
			segment := pulseSegment{
				FramesA:   nextTwoPulseFrames(g.config.OnHurtMode.PulseA.PulseWaveform, &pulseIndexA),
				FramesB:   nextTwoPulseFrames(g.config.OnHurtMode.PulseA.PulseWaveform, &pulseIndexB),
				StrengthA: g.config.OnHurtMode.StrengthA,
				StrengthB: g.config.OnHurtMode.StrengthB,
			}
			segmentList.PushBack(segment)
		}
		g.dequeLock.Lock()
		g.pulseDeque = list.New() //clear
		g.pulseDeque.PushFrontList(segmentList)
		g.dequeLock.Unlock()
		g.needContModeDecayCalc = true
	})

	_ = g.isaacListener.RegisterCallback(isaac2.PlayerDeathEvent, func(interface{}) {
		if !g.config.OnDeathMode.Enabled || !g.coyoteSession.IsBound() {
			return
		}

		var pulseIndexA int
		var pulseIndexB int
		segmentList := list.New()
		for duration := g.config.OnDeathMode.Duration; duration >= 0; duration -= 200 {
			segment := pulseSegment{
				FramesA:   nextTwoPulseFrames(g.config.OnDeathMode.PulseA.PulseWaveform, &pulseIndexA),
				FramesB:   nextTwoPulseFrames(g.config.OnDeathMode.PulseB.PulseWaveform, &pulseIndexB),
				StrengthA: g.config.OnDeathMode.StrengthA,
				StrengthB: g.config.OnDeathMode.StrengthB,
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

	_ = g.isaacListener.RegisterCallback(isaac2.ManualRestartEvent, func(callbackData interface{}) {
		if !g.config.OnManualRestart.Enabled || !g.coyoteSession.IsBound() {
			return
		}

		var pulseIndexA int
		var pulseIndexB int
		segmentList := list.New()
		for duration := g.config.OnManualRestart.Duration; duration >= 0; duration -= 200 {
			segment := pulseSegment{
				FramesA:   nextTwoPulseFrames(g.config.OnManualRestart.PulseA.PulseWaveform, &pulseIndexA),
				FramesB:   nextTwoPulseFrames(g.config.OnManualRestart.PulseB.PulseWaveform, &pulseIndexB),
				StrengthA: g.config.OnManualRestart.StrengthA,
				StrengthB: g.config.OnManualRestart.StrengthB,
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
	decayInterval := g.config.ContinuousMode.DecayInterval

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	var decayTicker *time.Ticker

	if decayInterval != 0 {
		decayTicker = time.NewTicker(time.Duration(decayInterval) * time.Millisecond)
		defer decayTicker.Stop()
	}

	var pulseIndexA int
	var pulseIndexB int
	var prevSegment pulseSegment

	for range ticker.C {
		if !g.config.ContinuousMode.Enabled || !g.coyoteSession.IsBound() || g.pulseDeque.Len() >= 100 {
			continue
		}

		if g.config.ContinuousMode.DecayInterval != decayInterval {
			if g.config.ContinuousMode.DecayInterval == 0 && decayTicker != nil {
				decayTicker.Stop()
			} else {
				decayInterval = g.config.ContinuousMode.DecayInterval
				decayTicker = time.NewTicker(time.Duration(decayInterval) * time.Millisecond)
			}
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
		healthDiff := g.playerInfo.MaxHealth - g.playerInfo.Health
		minA := g.config.ContinuousMode.BaseStrengthA + healthDiff*g.config.ContinuousMode.StrengthPerHealthA + g.collStrengthAddA
		minB := g.config.ContinuousMode.BaseStrengthB + healthDiff*g.config.ContinuousMode.StrengthPerHealthB + g.collStrengthAddB

		if g.config.ContinuousMode.DecayInterval > 0 {
			segment.StrengthA = prevSegment.StrengthA
			segment.StrengthB = prevSegment.StrengthB

			select {
			case <-decayTicker.C:
				if segment.StrengthA > minA {
					segment.StrengthA -= g.config.ContinuousMode.DecayValue
				}
				if segment.StrengthB > minB {
					segment.StrengthB -= g.config.ContinuousMode.DecayValue
				}
			default:
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

func NewGame(config *configModel.Game, coyoteSession *coyote.Session, isaacListener *isaac2.GameListener) *Game {
	return &Game{
		config:        config,
		coyoteSession: coyoteSession,
		isaacListener: isaacListener,

		pulseDeque: list.New(),
	}
}
