package game

import (
	"IsaacCoyote/pkg/coyote"
)

func nextTwoPulseFrames(pulse []coyote.PulseFrame, index *int) []coyote.PulseFrame {
	if len(pulse) == 0 {
		return []coyote.PulseFrame{}
	}
	*index %= len(pulse)
	frame0 := pulse[*index]

	nextIdx := (*index + 1) % len(pulse)
	frame1 := pulse[nextIdx]

	*index = (*index + 2) % len(pulse)
	return []coyote.PulseFrame{frame0, frame1}
}
