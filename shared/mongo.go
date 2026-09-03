package utils

import "time"

type MongoConfig struct {
	Uri                          string `mapstructure:"uri"`
	Database                     string `mapstructure:"database"`
	TimestampConversion          string `mapstructure:"timestamp_patterns"`
	UserTable                    string `mapstructure:"user_table"`
	ProfileTable                 string `mapstructure:"profile_table"`
	ProfileGroupTable            string `mapstructure:"profile_group_table"`
	ProfileRoleTable             string `mapstructure:"profile_role_table"`
	ProfilePermissionTable       string `mapstructure:"profile_permission_table"`
	ProfileDeviceTable           string `mapstructure:"profile_device_table"`
	ProfileDeviceGroupsTable     string `mapstructure:"profile_device_groups_table"`
	ProfileDeviceGroupsViewTable string `mapstructure:"profile_device_groups_view_table"`
	ResetPasswordTable           string `mapstructure:"reset_password_table"`

	ContactTable        string `mapstructure:"contact_table"`
	InstallTable        string `mapstructure:"install_table"`
	PlaceTable          string `mapstructure:"places_table"`
	ContactGroupTable   string `mapstructure:"contact_group_table"`
	DeviceTable         string `mapstructure:"device_table"`
	ReportTable         string `mapstructure:"report_table"`
	ReportScheduleTable string `mapstructure:"report_schedule_table"`
	AlertTable          string `mapstructure:"alert_table"`
	VasTable            string `mapstructure:"vas_table"`
	VasEventTable       string `mapstructure:"vas_event_table"`
	EmailTable          string `mapstructure:"email_table"`
	SmsTable            string `mapstructure:"sms_table"`
	GeofenceTable       string `mapstructure:"geofence_table"`
	GeofenceDeviceTable string `mapstructure:"geofence_device_table"`

	VasSubscriptionsTable string `mapstructure:"vas_subscriptions_table"`
	TasksTable            string `mapstructure:"tasks_table"`

	TransferTable                   string        `mapstructure:"transfer_table"`
	ReadTimeout                     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout                    time.Duration `mapstructure:"write_timeout"`
	ConnectTimeout                  time.Duration `mapstructure:"connect_timeout"`
	DeviceTypeTable                 string        `mapstructure:"device_type_table"`
	DeviceConfigurationVersionTable string        `mapstructure:"device_configuration_version_table"`
	DeviceConfigurationTable        string        `mapstructure:"device_configuration_table"`
	HistoryTable                    string        `mapstructure:"history_table"`
	TripTable                       string        `mapstructure:"trip_table"`
	LastTable                       string        `mapstructure:"last_table"`
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
