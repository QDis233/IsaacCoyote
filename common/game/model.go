package game

import "IsaacCoyote/pkg/coyote"

// pulseSegment 200ms
type pulseSegment struct {
	StrengthA int
	StrengthB int
	FramesA   []coyote.PulseFrame
	FramesB   []coyote.PulseFrame
}

type playerInfo struct {
	Health    int
	MaxHealth int
}
