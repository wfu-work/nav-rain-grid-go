package configs

import "strconv"

const DefaultMqttPort = 1883

type MqttConfig struct {
	Enable bool   `mapstructure:"enable" json:"enable" yaml:"enable"`
	Host   string `mapstructure:"host" json:"host" yaml:"host"`
	Port   int    `mapstructure:"port" json:"port" yaml:"port"`
}

func (c MqttConfig) Address() string {
	if c.Port == 0 {
		c.Port = DefaultMqttPort
	}
	if c.Host == "" {
		return ":" + strconv.Itoa(c.Port)
	}
	return c.Host + ":" + strconv.Itoa(c.Port)
}
