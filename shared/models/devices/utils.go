package devices

import (
	"encoding/json"

	utils "blueassetgroup.com/reports-service/shared"
)

type DeviceType struct {
	Name        string `json:"name" bson:"_id"`
	Description string `json:"description"`
}

type DeviceConfiguration struct {
	Name        string `json:"name" bson:"_id"`
	Description string `json:"description"`
}

type DeviceConfigurationVersion struct {
	Version     string `json:"version" bson:"_id"`
	Description string `json:"description"`
}

type DeviceTypeFindAllResult struct {
	utils.Result
	DeviceTypes []*DeviceType `json:"device_types"`
	Count       int           `json:"count"`
}

type DeviceConfigurationFindAllResult struct {
	utils.Result
	DeviceConfigurations []*DeviceConfiguration `json:"device_configurations"`
	Count                int                    `json:"count"`
}

type DeviceConfigurationVersionFindAllResult struct {
	utils.Result
	DeviceConfigurationVersions []*DeviceConfigurationVersion `json:"device_configuration_versions"`
	Count                       int                           `json:"count"`
}

type DeviceTypeAddRequest struct {
	Name        string `json:"name"  validate:"nonzero"`
	Description string `json:"description"  validate:"nonzero"`
}

func (u *DeviceTypeAddRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u DeviceTypeAddRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

type SetStatusRequest struct {
	IMEI      string `json:"imei"  validate:"nonzero"`
	Status    string `json:"status"  validate:"nonzero"`
	ProfileId string `json:"profile_id"`
}

func (u *SetStatusRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}
