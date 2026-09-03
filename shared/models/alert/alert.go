package alert

import (
	"encoding/json"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertStatus int

const (
	ALERT_STATUS_CREATED AlertStatus = 0
	ALERT_STATUS_CLOSED  AlertStatus = 1
	ALERT_STATUS_COMMENT AlertStatus = 2
)

func (c AlertStatus) String() string {
	switch c {
	case ALERT_STATUS_CREATED:
		return "CREATED"
	case ALERT_STATUS_CLOSED:
		return "CLOSED"
	case ALERT_STATUS_COMMENT:
		return "COMMENT"
	default:
		return "ERROR"
	}
}

type AlertEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Comment   string    `json:"comment"`
	Status    string    `json:"status"`
	UserName  string    `json:"user_name" validate:"nonzero"`
}

type AddAlertEventRequest struct {
	AlertId  string `json:"alert_id" validate:"nonzero"`
	Comment  string `json:"comment" validate:"nonzero"`
	UserName string `json:"user_name" validate:"nonzero"`
	Status   string `json:"status" validate:"nonzero"`
}

func (p AddAlertEventRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

func (u *AddAlertEventRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (r AddAlertEventRequest) ToBytes() []byte {
	b, _ := json.Marshal(r)
	return b
}

type Alert struct {
	ID         primitive.ObjectID `bson:"_id" json:"id"`
	Timestamp  time.Time          `json:"timestamp"`
	AlertType  string             `json:"alert_type" validate:"nonzero" bson:"alert_type"`
	IMEI       string             `json:"imei" validate:"nonzero"`
	Events     []AlertEvent       `json:"alert_event"`
	Profile    primitive.ObjectID `json:"profile_id"`
	DeviceName string             `json:"device_name" bson:"-"`
}

func (p Alert) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

func (u *Alert) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u Alert) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

type AlertAddResponse struct {
	utils.Result
	AlertId string `json:"id"`
}

type AlertFindAllResponse struct {
	Data  []*Alert `json:"alert"`
	Count int64    `json:"count"`
	utils.Result
}

func (u *AlertFindAllResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type AlertFindRequest struct {
	AlertId string `json:"alert_id"`
}

func (u *AlertFindRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (r AlertFindRequest) ToBytes() []byte {
	b, _ := json.Marshal(r)
	return b
}

type AlertFindResponse struct {
	utils.Result
	Alert *Alert `json:"alert"`
}

func (u *AlertFindResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}
