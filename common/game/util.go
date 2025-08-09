package game

import (
	"IsaacCoyote/common/isaac"
	"IsaacCoyote/pkg/coyote"
	"regexp"
	"strconv"
	"strings"
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

func parseCollectiblesString(s string, resManager *isaac.ResourceManager) ([]itemDetailWrapper, error) {
	re := regexp.MustCompile(`([^,:]+):([^,]+)`)
	matches := re.FindAllStringSubmatch(s, -1)

	result := make([]itemDetailWrapper, 0)
	for _, match := range matches {
		w := itemDetailWrapper{}
		if len(match) < 3 {
			continue
		}

		item, err := resManager.GetItemByName(strings.TrimSpace(match[1]))
		if err != nil {
			// Ignore items glitched without IDs or other information.
			continue
		}
		num, err := strconv.Atoi(strings.TrimSpace(match[2]))
		if err != nil {
			return nil, err
		}

		w.itemDetail = item
		w.num = num

		result = append(result, w)
	}

	return result, nil
}
