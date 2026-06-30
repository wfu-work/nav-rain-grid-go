package configs

import "strconv"

const DefaultPort = 1883

type MqttConfig struct {
	Enable bool   `mapstructure:"enable" json:"enable" yaml:"enable"`
	Host   string `mapstructure:"host" json:"host" yaml:"host"`
	Port   int    `mapstructure:"port" json:"port" yaml:"port"`
}

func (c MqttConfig) Address() string {
	if c.Port == 0 {
		c.Port = DefaultPort
	}
	if c.Host == "" {
		return ":" + strconv.Itoa(c.Port)
	}
	return c.Host + ":" + strconv.Itoa(c.Port)
}
