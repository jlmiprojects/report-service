package utils

type NatsConfig struct {
	Uri   string `mapstructure:"uri"`
	Token string `mapstructure:"token"`
}
