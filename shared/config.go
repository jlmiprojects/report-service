package utils

import (
	"strings"

	"blueassetgroup.com/reports-service/shared/config"
	"github.com/spf13/viper"
)

type Config struct {
	Port             int                         `mapstructure:"port"`
	Host             string                      `mapstructure:"host"`
	Mongo            MongoConfig                 `mapstructure:"mongo"`
	MongoConnections map[string]*MongoConnection `mapstructure:"mongo_connections"`
	Logger           *LoggerConfig               `mapstructure:"logger"`
	Nats             *NatsConfig                 `mapstructure:"nats"`
	Email            *config.EmailConfig         `mapstructure:"email"`

	TemplateDir *string `mapstructure:"template_dir"`
	StaticDir   *string `mapstructure:"static_dir"`
	ScriptDir   *string `mapstructure:"script_dir"`
	ChromeUrl   *string `mapstructure:"chrome_url"`
	ChromeHost  *string `mapstructure:"chrome_host"`
}

// NewConfig loads conf/application.json into a Config via viper, mirroring the
// email-service. It searches ./conf, ../conf, . and /etc so both the report
// service binary (run from its module root) and the scheduler resolve the same
// file regardless of the working directory.
//
// Environment variables override matching keys (viper AutomaticEnv). A config
// key maps to an env var by upper-casing it and joining the nesting levels with
// "_", which is what the SetEnvKeyReplacer call below provides -- without it
// viper looks for the literal name "MONGO.DATABASE" and no override ever fires.
// So:
//
//	mongo.database                      -> MONGO_DATABASE
//	mongo.uri                           -> MONGO_URI
//	mongo.report_table                  -> MONGO_REPORT_TABLE
//	mongo_connections.default.database  -> MONGO_CONNECTIONS_DEFAULT_DATABASE
//	nats.uri / nats.token               -> NATS_URI / NATS_TOKEN
//	port / host / chrome_url            -> PORT / HOST / CHROME_URL
//
// Two limits come with AutomaticEnv: only keys that are already present in
// application.json can be overridden (a key the file omits is invisible to
// viper), and mongo_connections overrides only reach connections the file
// declares; env cannot add a new named connection.
func NewConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigName("application")
	v.SetConfigType("json")
	v.AddConfigPath("./conf")
	v.AddConfigPath("../conf")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	c := new(Config)
	if err := v.Unmarshal(c); err != nil {
		return nil, err
	}
	return c, nil
}
