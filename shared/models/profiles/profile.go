package profiles

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	uploads "blueassetgroup.com/reports-service/shared/models/uploads"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FindProfilesResult struct {
	Profiles []*Profile `json:"profiles"`
	Count    int64      `json:"total_count"`
}

func (p *FindProfilesResult) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type FindAllRolesResult struct {
	utils.Result
	Roles []*Role `json:"roles"`
}

type FindAllProfileRequest struct {
	Params  string `json:"params"`
	Columns string `json:"columns"`
}

func (_self FindAllProfileRequest) ToJSON() ([]byte, error) {
	return json.Marshal(_self)
}

func (p *FindAllProfileRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type FindProfileRequest struct {
	ID       string `json:"id,omitempty"`
	UserName string `json:"user_name,omitempty"`
}

func (p *FindProfileRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (_self FindProfileRequest) ToJSON() ([]byte, error) {
	return json.Marshal(_self)
}

type FindProfileResponse struct {
	utils.Result
	Profile *Profile `json:"profile"`
}

func (p *FindProfileResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type FindProfileUploadsRequest struct {
	ID string `json:"id"`
}

func (p *FindProfileUploadsRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type UploadProfileImageRequest struct {
	ID         string `json:"id"`
	FileName   string `json:"file_name"`
	BucketName string `json:"bucket_name"`
}

func (p *UploadProfileImageRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type DeleteProfileUploadsRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (p *DeleteProfileUploadsRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type UploadResult struct {
	utils.Result
	Files []string `json:"files,omitempty"`
}

type CreateProfileResult struct {
	utils.Result
	ID string `json:"id"`
}

func (p *CreateProfileResult) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type Address struct {
	Line1 string `json:"line1,omitempty" bson:"line1,omitempty"`
	Line2 string `json:"line2,omitempty" bson:"line2,omitempty"`
	Line3 string `json:"line3,omitempty" bson:"line3,omitempty"`
	Line4 string `json:"line4,omitempty" bson:"line4,omitempty"`
}

type ProfileSettings map[string]interface{}

// login_expiry 10s

type Profile struct {
	ID       string             `json:"id" bson:"-"`
	MongoID  primitive.ObjectID `json:"-" bson:"_id"`
	UserName string             `json:"user_name,omitempty" bson:"user_name" validate:"nonzero"`
	Name     string             `json:"name,omitempty" bson:"name" validate:"nonzero"`
	Surname  string             `json:"surname,omitempty" bson:"surname" validate:"nonzero"`
	Email    string             `json:"email,omitempty" bson:"email" validate:"nonzero"`
	Password string             `json:"password" validate:"nonzero"`
	Comment  string             `json:"comment,omitempty" bson:"comment"`
	Language string             `json:"language,omitempty" bson:"language"`
	License  *License           `json:"license,omitempty" bson:"license,omitempty"`

	VatNo    string   `json:"vat_no" bson:"vat_no" `
	PastelNo string   `json:"pastel_no,omitempty" bson:"pastel_no,omitempty"`
	Status   string   `json:"status" bson:"status"`
	Id_Reg   string   `json:"id_reg,omitempty" bson:"id_reg,omitempty"`
	MobileNo string   `json:"mobile_no,omitempty" bson:"mobile_no,omitempty"`
	TelNo    string   `json:"tel_no,omitempty" bson:"tel_no,omitempty,omitempty"`
	Address  *Address `json:"address,omitempty" bson:"address,omitempty,omitempty"`

	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`
	LastLogin   time.Time `json:"last_login" bson:"last_login"`
	//CreatedAt   *time.Time `json:"created_at" bson:"created_at"`

	Avatar  uploads.Upload   `json:"avatar" bson:"avatar"`
	Uploads []uploads.Upload `json:"uploads,omitempty" bson:"uploads,omitempty"`

	Type    string `json:"type" bson:"type" validate:"nonzero"` // INDIVIDUAL | COMPANY
	Company string `json:"company,omitempty" bson:"company"`

	Settings *ProfileSettings `json:"settings,omitempty" bson:"settings,omitempty"`

	PremiumMaps           bool `json:"premium_maps" bson:"premium_maps"`
	Reports               bool `json:"reports" bson:"reports"`
	ReceiveSmsAlerts      bool `json:"alerts_sms" bson:"alerts_sms"`
	ReceiveEmailAlerts    bool `json:"alerts_email" bson:"alerts_email"`
	ReceiveWhatsAppAlerts bool `json:"alerts_wa" bson:"alerts_wa"`
}

func (u Profile) GetUserNameField() string {
	return "user_name"
}

func (p *Profile) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p Profile) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

func (u Profile) GetAllBSONFieldsNames() []string {

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

type Role struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Permissions []string           `json:"permissions"`
}

type Permission struct {
	ID          string `json:"id" bson:"_id"`
	Description string `json:"description"`
}

type SubProfile struct {
	Id           primitive.ObjectID `bson:"_id" json:"id"`
	ProfileId    primitive.ObjectID `json:"-" bson:"profile_id"`
	Profile      *Profile           `json:"profile,omitempty" bson:"-"`
	SubProfileId primitive.ObjectID `json:"-" bson:"sub_profile_id"`
	SubProfile   *Profile           `json:"sub_profile" bson:"-"`
}

type SubProfileAddRequest struct {
	Parent   string   `json:"profile_id"`
	Children []string `json:"sub_profiles"`
}

func (p *SubProfileAddRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (_self SubProfileAddRequest) ToJSON() ([]byte, error) {
	return json.Marshal(_self)
}

type SubProfileFindAllRequest struct {
	Profile string `json:"profile_id"`
}

func (p *SubProfileFindAllRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type SubProfileFindAllResponse struct {
	utils.Result
	Data []*Profile `json:"sub_profiles"`
}

type SubProfileRemoveRequest struct {
	ProfileId   string   `json:"profile_id"`
	SubProfiles []string `json:"sub_profiles"`
}

func (p *SubProfileRemoveRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type ProfileDevice struct {
	ProfileId primitive.ObjectID `json:"profile_id" validate:"nonzero"`
	IMEI      string             `json:"imei" validate:"nonzero"`
	IMEIS     []string           `json:"imeis"`
}

func (p *ProfileDevice) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (_self ProfileDevice) ToJSON() ([]byte, error) {
	return json.Marshal(_self)
}

type ProfileDeviceFindAllResponse struct {
	utils.Result
	Data []*ProfileDevice `json:"data"`
}

func (p *ProfileDeviceFindAllResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p ProfileDevice) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

type ProfileDeviceAddResponse struct {
	utils.Result
	ID string `json:"id"`
}

type ProfileDeviceFindAllRequest struct {
	ProfileId string `json:"profile_id" validate:"nonzero"`
	GroupId   string `json:"group_id"`
}

func (_self ProfileDeviceFindAllRequest) ToJSON() ([]byte, error) {
	return json.Marshal(_self)
}

func (p ProfileDeviceFindAllRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

func (p *ProfileDeviceFindAllRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (_self Profile) ToJSON() ([]byte, error) {
	return json.Marshal(_self)
}
