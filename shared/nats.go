package utils

type NatsConfig struct {
	Uri     string `mapstructure:"uri"`
	Token   string `mapstructure:"token"`
	Subject string `mapstructure:"subject"`
	Queue   string `mapstructure:"queue"`
}

type JetStreamConfig struct {
	Durable string `json:"durable"`
	Stream  string `json:"stream"`
	Filter  string `json:"filter"`
}
