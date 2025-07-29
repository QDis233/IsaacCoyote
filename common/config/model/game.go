package model

import (
	"IsaacCoyote/pkg/coyote"
)

type StrengthOperator string

const (
	INCREMENT StrengthOperator = "INCREMENT"
	ABSOLUTE  StrengthOperator = "ABSOLUTE"
)

func (s *StrengthOperator) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var opString string
	if err := unmarshal(&opString); err != nil {
		return err
	}
	*s = StrengthOperator(opString)
	return nil
}

type PulseConfig struct {
	PulseWaveform coyote.PulseWaveform
	rawData       string
}

func (p *PulseConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var pulseString string
	if err := unmarshal(&pulseString); err != nil {
		return err
	}
	pw, err := coyote.UnmarshalPulseFromString(pulseString)
	if err != nil {
		return err
	}
	p.PulseWaveform = pw
	p.rawData = pulseString
	return nil
}

type Game struct {
	BaseStrengthA      int `yaml:"base_strength_A"`
	BaseStrengthB      int `yaml:"base_strength_B"`
	StrengthPerHealthA int `yaml:"strength_per_health_A"`
	StrengthPerHealthB int `yaml:"strength_per_health_B"`

	ContinuousMode   ContinuousMode   `yaml:"continuous_mode"`
	OnNewCollectible OnNewCollectible `yaml:"on_new_collectible"`
	OnHurt           OnHurt           `yaml:"on_hurt"`
	OnDeath          OnDeath          `yaml:"on_death"`
	OnManualRestart  OnManualRestart  `yaml:"on_manual_restart"`
}

type ContinuousMode struct {
	Enabled bool `yaml:"enabled"`

	DecayInterval int `yaml:"decay_interval"`
	DecayValue    int `yaml:"decay_value"`

	PulseA PulseConfig `yaml:"pulse_A"`
	PulseB PulseConfig `yaml:"pulse_B"`
}

type OnNewCollectible struct {
	Enabled        bool `yaml:"enabled"`
	StrengthConfig map[int]struct {
		StrengthAddA int `yaml:"strength_add_A"`
		StrengthAddB int `yaml:"strength_add_B"`
	} `yaml:"strength_config"`
}

type OnHurt struct {
	Enabled bool `yaml:"enabled"`

	Duration int `yaml:"duration"`

	StrengthOperator StrengthOperator `yaml:"strength_operator"`
	StrengthA        int              `yaml:"strength_A"`
	StrengthB        int              `yaml:"strength_B"`

	PulseA PulseConfig `yaml:"pulse_A"`
	PulseB PulseConfig `yaml:"pulse_B"`
}

type OnDeath struct {
	Enabled bool `yaml:"enabled"`

	Duration int `yaml:"duration"`

	StrengthOperator StrengthOperator `yaml:"strength_operator"`
	StrengthA        int              `yaml:"strength_A"`
	StrengthB        int              `yaml:"strength_B"`

	PulseA PulseConfig `yaml:"pulse_A"`
	PulseB PulseConfig `yaml:"pulse_B"`
}

type OnManualRestart struct {
	Enabled  bool `yaml:"enabled"`
	Duration int  `yaml:"duration"`

	StrengthOperator StrengthOperator `yaml:"strength_operator"`
	StrengthA        int              `yaml:"strength_A"`
	StrengthB        int              `yaml:"strength_B"`

	PulseA PulseConfig `yaml:"pulse_A"`
	PulseB PulseConfig `yaml:"pulse_B"`
}
