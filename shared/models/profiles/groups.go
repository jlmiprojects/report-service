package profiles

import (
	"encoding/json"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProfileDeviceGroup struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name"`
	ProfileId primitive.ObjectID `json:"profile_id"`
	Devices   any                `json:"devices"`
}

type ProfileDeviceGroupAddRequest struct {
	GroupId   string   `json:"id"`
	Name      string   `json:"name" validate:"nonzero"`
	ProfileId string   `json:"profile_id" validate:"nonzero"`
	Devices   []string `json:"devices"`
}

func (u *ProfileDeviceGroupAddRequest) FromJSON(data []byte) error {
	err := json.Unmarshal(data, u)

	if err != nil {
		return nil
	}

	return nil
}

func (u ProfileDeviceGroupAddRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

type ProfileDeviceGroupAddResponse struct {
	utils.Result
	ID string `json:"id"`
}

type ProfileDeviceGroupFindRequest struct {
	ProfileId   string `json:"profile_id" validate:"nonzero"`
	GroupId     string `json:"group_id"`
	ShowDetails bool   `json:"details"`
}

func (_self ProfileDeviceGroupFindRequest) ToJSON() ([]byte, error) {
	return json.Marshal(_self)
}

func (u *ProfileDeviceGroupFindRequest) FromJSON(data []byte) error {
	err := json.Unmarshal(data, u)
	if err != nil {
		return nil
	}
	return nil
}

type ProfileDeviceGroupFindResponse struct {
	utils.Result
	Groups []*ProfileDeviceGroup `json:"groups"`
}

func (u *ProfileDeviceGroupFindResponse) FromJSON(data []byte) error {
	err := json.Unmarshal(data, u)
	if err != nil {
		return nil
	}
	return nil
}

type ProfileDeviceGroupRemoveRequest struct {
	ID string `json:"id"`
}

func (u *ProfileDeviceGroupRemoveRequest) FromJSON(data []byte) error {
	err := json.Unmarshal(data, u)
	if err != nil {
		return nil
	}
	return nil
}

type ForgetPasswordRequest struct {
	UserName string `json:"user_name" validate:"nonzero"`
}

func (u *ForgetPasswordRequest) FromJSON(data []byte) error {
	err := json.Unmarshal(data, u)
	if err != nil {
		return nil
	}
	return nil
}

type ResetPassword struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	UserName  string             `json:"user_name" bson:"user_name"`
	ProfileId primitive.ObjectID `json:"profile_id" bson:"profile_id"`
	Timestamp time.Time          `json:"timestamp" bson:"timestamp"`
}

type ResetPassword2 struct {
	UUID     string `json:"uuid" `
	Password string `json:"password"`
}

func (u *ResetPassword2) FromJSON(data []byte) error {
	err := json.Unmarshal(data, u)
	if err != nil {
		return nil
	}
	return nil
}

type ForgetPasswordResponse struct {
	utils.Result
	UUID string `json:"uuid" `
}

type ChangePassword2Request struct {
	UUID           string `json:"uuid"`
	Password       string `json:"password"`
	PasswordRepeat string `json:"password_repeat"`
}
