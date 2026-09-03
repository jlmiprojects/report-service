package sms

import (
	"encoding/json"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Sms struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	To        string             `json:"to"`
	Message   string             `json:"message"`
	Status    string             `json:"status"`
	Sequence  int32              `json:"sequence"`
	MessageID string             `json:"message_id"`
	Created   time.Time          `json:"created"`
	Updated   time.Time          `json:"updated"`
	Data      any                `json:"data"`
	//IMEI      string             `json:"imei"`
	//ProfileId primitive.ObjectID `json:"profile_id" bson:"profile_id"`
}

type SmsSendRequest struct {
	To       []string `json:"to" validate:"nonzero"`
	Template string   `json:"template" validate:"nonzero"`
	Data     any      `json:"data"`
}

func (p SmsSendRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

func (u *SmsSendRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u SmsSendRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

type WASendRequest struct {
	To       string `json:"to" validate:"nonzero"`
	Template string `json:"template" validate:"nonzero"`
	Data     any    `json:"data"`
}

func (p WASendRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

func (u *WASendRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u WASendRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}
