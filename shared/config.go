package utils

import (
	"time"

	"blueassetgroup.com/reports-service/shared/config"
	"github.com/spf13/viper"
)

type Config struct {
	Port             int                         `mapstructure:"port"`
	Host             string                      `mapstructure:"host"`
	Mongo            MongoConfig                 `mapstructure:"mongo"`
	MongoConnections map[string]*MongoConnection `mapstructure:"mongo_connections"`
	Secret           string                      `mapstructure:"secret"`
	UploadPath       *string                     `mapstructure:"upload_path" `
	ImportPath       *string                     `mapstructure:"import_path"`
	Logger           *LoggerConfig               `mapstructure:"logger"`
	Nats             *NatsConfig                 `mapstructure:"nats"`
	ReversLookupUrl  *string                     `mapstructure:"reverse_lookup_url"`
	ResetPasswordUrl *string                     `mapstructure:"reset_password_url"`
	Sms              *SmsConfig                  `mapstructure:"sms"`
	Email            *config.EmailConfig         `mapstructure:"email"`
	WhatsApp         *config.WhatsAppConfig      `mapstructure:"whatsapp"`

	TemplateDir  *string           `mapstructure:"template_dir"`
	StaticDir    *string           `mapstructure:"static_dir"`
	ScriptDir    *string           `mapstructure:"script_dir"`
	JetStream    *JetStreamConfig  `mapstructure:"jetstream"`
	ChromeUrl    *string           `mapstructure:"chrome_url"`
	Vas          *config.VasConfig `mapstructure:"vas"`
	RedisConfig  *RedisConfig      `mapstructure:"redis"`
	DefaultTtl   time.Duration     `mapstructure:"default_ttl"`
	HRGUrl       *string           `mapstructure:"hrg_url"` // HRG is for road info lookups
	NumberFields *string           `mapstructure:"number_fields"`
}

// NewConfig loads conf/application.json into a Config via viper, mirroring the
// email-service. It searches ./conf, ../conf, . and /etc so both the report
// service binary (run from its module root) and the scheduler resolve the same
// file regardless of the working directory. Environment variables override
// matching keys (viper AutomaticEnv).
func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("application")
	v.SetConfigType("json")
	v.AddConfigPath("./conf")
	v.AddConfigPath("../conf")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	c := new(Config)
	if err := v.Unmarshal(c); err != nil {
		return nil, err
	}
	return c, nil
}
