package coyote

import (
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strconv"
)

// ParseStrengthData
// return [StrengthA, StrengthB, MaxStrengthA,MaxStrengthB]
func ParseStrengthData(strengthData string) ([4]int, error) {
	result := [4]int{}

	re, err := regexp.Compile(`^strength-(\d+)\+(\d+)\+(\d+)\+(\d+)$`)
	if err != nil {
		return result, err
	}
	match := re.FindStringSubmatch(strengthData)
	if len(match) != 5 {
		return result, InvalidMessageError{Message: "Invalid Strength Data"}
	}
	for i := 1; i < 5; i++ {
		strength, err := strconv.Atoi(match[i])
		if err != nil {
			return result, InvalidMessageError{Message: "Invalid Strength Data"}
		}
		result[i-1] = strength
	}

	return result, nil
}

// ParseFeedbackData
// return button index
// A:0,1,2,3,4 | B:5,6,7,8,9 (from left to right)
func ParseFeedbackData(feedbackData string) (int, error) {
	re, err := regexp.Compile(`^feedback-(\d+)$`)
	if err != nil {
		return 0, err
	}
	match := re.FindStringSubmatch(feedbackData)
	if len(match) != 2 {
		return 0, InvalidMessageError{Message: "Invalid Feedback Data"}
	}
	return strconv.Atoi(match[1])
}

// UnmarshalPulseWaveform
// Parse from hex to PulseWaveform
func UnmarshalPulseWaveform(pulse []string) (PulseWaveform, error) {
	pulseWaveform := make(PulseWaveform, len(pulse))
	for i, frame := range pulse {
		pulseFrame := PulseFrame{}
		frameData, err := hex.DecodeString(frame)
		if err != nil {
			return nil, err
		}
		if len(frameData) != 8 {
			return nil, InvalidPulseParamError{
				Message: "Invalid Pulse Frame Length",
			}
		}

		for j := 0; j < 4; j++ {
			pulseFrame.FrequencyData[j] = int(frameData[j])
		}
		for j := 4; j < 8; j++ {
			pulseFrame.StrengthData[j-4] = int(frameData[j])
		}

		pulseWaveform[i] = pulseFrame
	}
	return pulseWaveform, nil
}

func UnmarshalPulseFromString(pulseString string) (PulseWaveform, error) {
	if pulseString == "" || pulseString == "[]" {
		return PulseWaveform{}, nil
	}
	var data []string
	err := json.Unmarshal([]byte(pulseString), &data)
	if err != nil {
		return nil, err
	}
	return UnmarshalPulseWaveform(data)
}
