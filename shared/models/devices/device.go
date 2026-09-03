package devices

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"blueassetgroup.com/reports-service/shared/models/realtime"
	"blueassetgroup.com/reports-service/shared/models/uploads"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LinkedProfile struct {
	ID      string             `json:"id" bson:"-"`
	MongoID primitive.ObjectID `json:"-" bson:"id"`
}

const (
	TrackerTypeVehicle  = "VEHICLE"
	TrackerTypePersonal = "PERSONAL"
	TrackerTypeAsset    = "ASSET"
)

const (
	DeviceStatusActive    = "active"
	DeviceStatusSuspended = "suspended"
	DeviceStatusDeleted   = "deleted"
)

type Device struct {
	IMEI          string `json:"imei" bson:"imei" validate:"nonzero,min=15,max=15"`
	Name          string `json:"name" bson:"name" validate:"nonzero,min=5"`
	Msisdn        string `json:"msisdn" bson:"msisdn" validate:"nonzero,min=10"`
	DeviceType    string `json:"device_type" bson:"device_type" validate:"nonzero"`
	Configuration string `json:"configuration" validate:"nonzero" csv:"configuration"`
	TrackerType   string `json:"tracker_type" bson:"tracker_type" validate:"nonzero"`
	Version       string `json:"version" validate:"nonzero" csv:"version"`
	VIN           string `json:"vin" bson:"vin"`
	EngineNumber  string `json:"engine_number" bson:"engine_number"`

	Description           string `json:"description" bson:"description"`
	IconName              string `json:"icon_name" bson:"icon_name"`
	IconColour            string `json:"icon_colour" bson:"icon_colour"`
	Make                  string `json:"make" bson:"make"`
	Model                 string `json:"model" bson:"model"`
	Ignition              bool   `json:"ignition" bson:"ignition"`
	Colour                string `json:"colour" bson:"colour"`
	LongTermPark          bool   `json:"long_term_park" bson:"long_term_park"`
	ServiceInternal       int    `json:"service_interval" bson:"service_interval"`
	LicenceExpiry         string `json:"licence_expiry" bson:"licence_expiry"`
	LicenceExpiryReminder string `json:"licence_expiry_reminder" bson:"licence_expiry_reminder"`

	RegularDriver primitive.ObjectID `json:"regular_driver"`
	Driver        string             `json:"driver"`

	Comment string `json:"comment" bson:"comment"`

	Owner       primitive.ObjectID `json:"owner" bson:"owner"`
	OwnerDetail map[string]any     `json:"owner_detail,omitempty" bson:"-"`

	PrimaryContact primitive.ObjectID `json:"primary_contact" bson:"primary_contact"`

	SecondaryContact primitive.ObjectID `json:"secondary_contact" bson:"secondary_contact"`

	LastUpdated time.Time `json:"updated_at" bson:"updated_at"`
	UpdatedBy   string    `json:"updated_by" bson:"updated_by"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	CreatedBy string    `json:"created_by" bson:"created_by"`

	Status string `json:"status" bson:"status"`

	Image uploads.Upload `json:"image" bson:"image"`

	Uploads []uploads.Upload `json:"uploads" bson:"uploads,omitempty"`

	VAS []string `json:"vas" bson:"-"`

	// Extra fields for stock
	StockStatus       string            `json:"stock_status" bson:"stock_status"`
	FitmentCentre     string            `json:"fitment_centre" bson:"fitment_centre"`
	FitmentCentrePoNo string            `json:"fitness_centre_po_no" bson:"fitness_centre_po_no"`
	WaybillNo         string            `json:"waybill_no" bson:"waybill_no"`
	DateInstalled     *time.Time        `json:"installed_date" bson:"installed_date"`
	LocationInstalled string            `json:"installed_location" bson:"installed_location"`
	LastPoint         *realtime.Message `json:"last_point" bson:"-"`
}

func (u Device) GetAllBSONFieldsNames() []string {

	retval := make([]string, 0)
	val := reflect.ValueOf(u)
	for i := 0; i < val.Type().NumField(); i++ {

		if len(val.Type().Field(i).Tag.Get("bson")) > 0 && val.Type().Field(i).Tag.Get("bson") != "-" {

			parts := strings.Split(val.Type().Field(i).Tag.Get("bson"), ",")

			retval = append(retval, parts[0])
		}
	}

	return retval

}

func (u Device) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

func (u *Device) FromJSON(data []byte) error {
	err := json.Unmarshal(data, u)

	if err != nil {
		return nil
	}

	return nil
}

type FindDevicesResult struct {
	utils.Result
	Devices []*Device `json:"devices"`
	Count   int64     `json:"total_count,omitempty"`
}

func (u *FindDevicesResult) FromJSON(data []byte) error {
	err := json.Unmarshal(data, u)

	if err != nil {
		return nil
	}

	return nil
}

type FindDeviceRequest struct {
	IMEI string `json:"imei"`
	VAS  any    `json:"vas"`
}

func (u FindDeviceRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *FindDeviceRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type FindDeviceResult struct {
	utils.Result
	Device *Device `json:"device,omitempty"`
}

func (u *FindDeviceResult) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type UploadDeviceResult struct {
	utils.Result
	Errors []string
}

type ImportDevicesRequest struct {
	FileName   string `json:"file_name"`
	BucketName string `json:"bucket_name"`
}

func (u *ImportDevicesRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type GeoCodeRequest struct {
	Lat any `json:"lat"`
	Lon any `json:"lon"`
}

func (u *GeoCodeRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type GeoCodeResponse struct {
	utils.Result
	Address string `json:"address"`
}

type Place struct {
	Error   string `json:"error"`
	Address string `json:"display_name"`
}

type ProvisionRequest struct {
	Imei   string `json:"imei" validate:"nonzero,min=15" csv:"imei"`
	Msisdn string `json:"msisdn" csv:"msisdn"`

	DeviceType    string `json:"device_type" validate:"nonzero" csv:"device_type"`
	Configuration string `json:"configuration" validate:"nonzero" csv:"configuration"`
	Version       string `json:"version" validate:"nonzero" csv:"version"`

	Make        string `json:"make" csv:"make"`
	Name        string `json:"name" csv:"name"`
	Description string `json:"description" csv:"description"`
	Comment     string `json:"comment" csv:"comment"`
	Model       string `json:"model" csv:"model"`
}

func (u ProvisionRequest) ToJSON() []byte {
	b, _ := json.Marshal(u)
	return b
}

type ProvisionResponse struct {
	utils.Result
	Imei string `json:"imei"`
}

func (u *ProvisionResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u *ProvisionRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u ProvisionRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

type SetOdoRequest struct {
	IMEI string `json:"imei"`
	Odo  int64  `json:"odo"`
}

func (u *SetOdoRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type ImmobilizeRequest struct {
	IMEI  string `json:"imei"`
	Value int    `json:"value"`
}

func (u *ImmobilizeRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type UpdateOneRequest struct {
	IMEI  string   `json:"imei"`
	IMEIS []string `json:"imeis"`
	Name  string   `json:"name"`
	Value any      `json:"value"`
}

func (u UpdateOneRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *UpdateOneRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type AssignOwnerRequest struct {
	IMEI      []string `json:"imei"`
	ProfileId string   `json:"profile_id"`
}

func (u *AssignOwnerRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u AssignOwnerRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

type PanicRequest struct {
	IMEI      string `json:"imei"`
	ProfileId string `json:"profile_id"`
}

func (u *PanicRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}
