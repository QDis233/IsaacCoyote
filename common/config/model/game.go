package model

import (
	"IsaacCoyote/pkg/coyote"
)

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
	ContinuousMode  ContinuousMode  `yaml:"continuous_mode"`
	OnHurtMode      OnHurtMode      `yaml:"on_hurt_mode"`
	OnDeathMode     OnDeathMode     `yaml:"on_death_mode"`
	OnManualRestart OnManualRestart `yaml:"on_manual_restart"`
}

type ContinuousMode struct {
	Enabled bool `yaml:"enabled"`

	DecayInterval int `yaml:"decay_interval"`
	DecayValue    int `yaml:"decay_value"`

	BaseStrengthA int `yaml:"base_strength_A"`
	BaseStrengthB int `yaml:"base_strength_B"`

	StrengthPerHealthA int `yaml:"strength_per_health_A"`
	StrengthPerHealthB int `yaml:"strength_per_health_B"`

	OnNewCollectible OnNewCollectible `yaml:"on_new_collectible"`

	PulseA PulseConfig `yaml:"pulse_A"`
	PulseB PulseConfig `yaml:"pulse_B"`
}

type OnNewCollectible struct {
	Enabled        bool `yaml:"enabled"`
	StrengthConfig map[int]struct {
		StrengthA int `yaml:"strength_A"`
		StrengthB int `yaml:"strength_B"`
	} `yaml:"strength_config"`
}

type OnHurtMode struct {
	Enabled bool `yaml:"enabled"`

	Duration int `yaml:"duration"`

	StrengthA int `yaml:"strength_A"`
	StrengthB int `yaml:"strength_B"`

	PulseA PulseConfig `yaml:"pulse_A"`
	PulseB PulseConfig `yaml:"pulse_B"`
}

type OnDeathMode struct {
	Enabled bool `yaml:"enabled"`

	Duration int `yaml:"duration"`

	StrengthA int `yaml:"strength_A"`
	StrengthB int `yaml:"strength_B"`

	PulseA PulseConfig `yaml:"pulse_A"`
	PulseB PulseConfig `yaml:"pulse_B"`
}

type OnManualRestart struct {
	Enabled   bool        `yaml:"enabled"`
	Duration  int         `yaml:"duration"`
	StrengthA int         `yaml:"strength_A"`
	StrengthB int         `yaml:"strength_B"`
	PulseA    PulseConfig `yaml:"pulse_A"`
	PulseB    PulseConfig `yaml:"pulse_B"`
}
