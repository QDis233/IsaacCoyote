package model

type ConfigRoot struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Debug   bool   `yaml:"debug"`

	Coyote Coyote `yaml:"coyote"`
	Game   Game   `yaml:"game"`
}
