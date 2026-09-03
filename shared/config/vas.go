package config

import "time"

type VasConfig struct {
	EventSubject string `mapstructure:"event_subject"`

	SpeedName        string        `mapstructure:"speed_name"`
	SpeedAlertExpiry time.Duration `mapstructure:"speed_expiry"`

	MovementName        string        `mapstructure:"movement_name"`
	MovementAlertExpiry time.Duration `mapstructure:"movement_expiry"`

	PanicName        string        `mapstructure:"panic_name"`
	PanicAlertExpiry time.Duration `mapstructure:"panic_expiry"`

	TamperName        string        `mapstructure:"tamper_name"`
	TamperAlertExpiry time.Duration `mapstructure:"tamper_expiry"`

	GeofenceName        string        `mapstructure:"geofence_name"`
	GeofenceAlertExpiry time.Duration `mapstructure:"geofence_expiry"`

	VasEventConfigs map[string]GenericVasEventConfig `mapstructure:"events"`

	//JammingConfig GenericConfig `mapstructure:"jamming"`

	SvrName    string `mapstructure:"svr_name"`
	SvrEnabled bool   `mapstructure:"svr_enabled"`
}

type GenericVasEventConfig struct {
	Name        string        `mapstructure:"name"`
	AlertExpiry time.Duration `mapstructure:"expiry"`
	Recovery    bool          `mapstructure:"recovery"`
}
