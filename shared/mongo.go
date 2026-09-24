package utils

import "time"

// MongoConfig is the service's own database: where the report definitions and
// schedules live (the `mongo` block in application.json).
type MongoConfig struct {
	Uri                 string        `mapstructure:"uri"`
	Database            string        `mapstructure:"database"`
	ReportTable         string        `mapstructure:"report_table"`
	ReportScheduleTable string        `mapstructure:"report_schedule_table"`
	ConnectTimeout      time.Duration `mapstructure:"connect_timeout"`
	ReadTimeout         time.Duration `mapstructure:"read_timeout"`
}

// MongoConnection is one named datasource a `mongo` report DataAction can query
// against. They are configured under `mongo_connections` in application.json and
// selected per action by name (blank => "default").
type MongoConnection struct {
	Uri            string        `mapstructure:"uri"`
	Database       string        `mapstructure:"database"`
	ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
}
