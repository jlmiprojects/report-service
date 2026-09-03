package utils

import "time"

type SmsConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	SystemID       string        `mapstructure:"system_id"`
	SystemType     string        `mapstructure:"system_type"`
	Password       string        `mapstructure:"password"`
	RebindInterval time.Duration `mapstructure:"rebind_interval"`
}
