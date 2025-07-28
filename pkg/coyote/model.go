package coyote

import (
	"IsaacCoyote/pkg/coyote/enums"
	"encoding/hex"
	"encoding/json"
	"strings"
)

type WSMessage struct {
	Type     enums.MsgType `json:"type"`
	ClientID string        `json:"clientId"`
	TargetID string        `json:"targetId"`
	MsgData  string        `json:"message"`
}

func (m *WSMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

type StrengthData struct {
	StrengthA    int
	StrengthB    int
	MaxStrengthA int
	MaxStrengthB int
}

type CallbackData[T any] struct {
	CallbackData T
	Message      *WSMessage
}

type PulseFrame struct {
	StrengthData  [4]int
	FrequencyData [4]int
}

func (p *PulseFrame) Marshal() (string, error) {
	data := make([]byte, 8)

	for i, freq := range p.FrequencyData {
		if freq < 10 || freq > 240 {
			return "", InvalidPulseParamError{
				Message: "Invalid frequency (10,240)",
			}
		}
		data[i] = byte(freq)
	}

	for i, strength := range p.StrengthData {
		if strength < 0 || strength > 100 {
			return "", InvalidPulseParamError{
				Message: "Invalid strength value (0,100)",
			}
		}
		data[4+i] = byte(strength)
	}

	return strings.ToUpper(hex.EncodeToString(data)), nil
}

type PulseWaveform []PulseFrame

func (p *PulseWaveform) Marshal() ([]string, error) {
	var result []string

	if len(*p) > 100 {
		return nil, TooLongPulseError{
			Message: "pulse length larger than 100",
		}
	}

	for _, frame := range *p {
		frameData, err := frame.Marshal()
		if err != nil {
			return nil, err
		}
		result = append(result, frameData)
	}

	return result, nil
}
