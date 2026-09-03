package vas

import (
	"encoding/json"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Vas struct {
	Name         string    `json:"name" bson:"_id" validate:"nonzero"`
	Description  string    `json:"description" bson:"description" validate:"nonzero"`
	DeviceTypes  []string  `json:"device_types" bson:"device_types" validate:"nonzero"`
	LastUpdated  time.Time `json:"last_updated" bson:"last_updated"`
	Alerts       bool      `json:"alerts" bson:"alerts"`
	Cost         float32   `json:"cost" bson:"cost"`
	Configurable bool      `json:"configurable" bson:"configurable"`
}

func (o *Vas) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

func (u Vas) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

type VasFindAllRequest struct {
	DeviceType string `json:"device_type"`
	AlertsOnly any    `json:"alerts"`
}

func (o *VasFindAllRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

type VasFindAllResponse struct {
	Data  []*Vas `json:"vas"`
	Count int    `json:"count"`
	utils.Result
}

type VasEvent struct {
	ID        primitive.ObjectID `json:"-" bson:"_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	IMEI      string             `json:"imei" bson:"imei"`
	ProfileId primitive.ObjectID `json:"profile_id" bson:"profile_id"`
	Vas       string             `json:"vas" bson:"vas"`
	Data      map[string]any     `json:"data" bson:"data"`
}

func (o *VasEvent) FromJson(data []byte) error {
	return json.Unmarshal(data, o)
}

func (vasevent *VasEvent) ToJson() ([]byte, error) {
	return json.Marshal(vasevent)
}

type VasSubscription struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	VAS            string             `json:"vas" validate:"nonzero"`
	IMEI           string             `json:"imei" validate:"nonzero"`
	IMEIS          []string           `json:"imeis" validate:"nonzero"`
	ProfileId      string             `json:"profile_id" bson:"-" validate:"nonzero"`
	MongoProfileId primitive.ObjectID `json:"-" bson:"profile_id"`
	Settings       interface{}        `json:"settings" bson:"settings"`
	LastUpdated    time.Time          `json:"last_updated" bson:"last_updated"`
	Alerts         bool               `json:"alerts" bson:"-"`
	//ßßNotifications:
}

func (o *VasSubscription) FromJSON(data []byte) error {

	err := json.Unmarshal(data, o)
	if err != nil {
		return err
	}

	return nil
}

func (u VasSubscription) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

type VasSubscriptionsRequest struct {
	IMEI          string   `json:"imei" validate:"nonzero"`
	ProfileId     string   `json:"profile_id" bson:"-" validate:"nonzero"`
	Subscriptions []string `json:"subscriptions" `
}

func (o *VasSubscriptionsRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

func (u VasSubscriptionsRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u VasSubscriptionsRequest) Validate() []*utils.ValidationError {

	return utils.GetValidate().Validate(u)
}

type VasFindAllSubscriptionsRequest struct {
	IMEI      string   `json:"imei"`
	ProfileId string   `json:"profile_id"`
	IMEIS     []string `json:"imeis"`
}

func (u VasFindAllSubscriptionsRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (o *VasFindAllSubscriptionsRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

type VasFindAllSubscriptionsResponse struct {
	utils.Result
	Data  []*VasSubscription `json:"subscriptions"`
	Count int                `json:"count"`
}

func (o *VasFindAllSubscriptionsResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

type Notification struct {
	Name string
}

type NotficationContact struct {
	ID    primitive.ObjectID `json:"contact_id" bson:"contact_id" validate:"nonzero" mapstructure:"contact_id"`
	Email bool               `json:"email" mapstructure:"email"`
	Sms   bool               `json:"sms" mapstructure:"sms"`
}

type Settings struct {
	Active   bool                 `json:"active" mapstructure:"active"`
	Contacts []NotficationContact `json:"contacts" mapstructure:"contacts"`
}

type SpeedSettings struct {
	Active   bool                 `json:"active" bson:"active"`
	Contacts []NotficationContact `json:"contacts" mapstructure:"contacts"`
	Speed40  bool                 `json:"speed_40" mapstructure:"speed_40"`
	Speed60  bool                 `json:"speed_60" mapstructure:"speed_60"`
	Speed80  bool                 `json:"speed_80" mapstructure:"speed_80"`
	Speed100 bool                 `json:"speed_100" mapstructure:"speed_100"`
	Speed120 bool                 `json:"speed_120" mapstructure:"speed_120"`
}

func (u SpeedSettings) Validate() []*utils.ValidationError {

	return utils.GetValidate().Validate(u)
}

type MovementSettings struct {
	Active   bool
	Contacts []NotficationContact `json:"contacts" mapstructure:"contacts"`
}

type BatterySettings struct {
	Active     bool
	Contacts   []NotficationContact `json:"contacts" mapstructure:"contacts"`
	Percentage float32              `json:"percentage" validate:"nonzero"`
}

func (u BatterySettings) Validate() []*utils.ValidationError {

	return utils.GetValidate().Validate(u)
}

type GeoFenceSettings struct {
	Fences []struct {
		Active   bool
		Contacts []NotficationContact `json:"contacts" mapstructure:"contacts"`
		AlertOn  string               `json:"alert_on" bson:"alert_on"`
		FenceId  primitive.ObjectID   `json:"fence_id" bson:"fence_id"`
	} `json:"fences" bson:"fences"`
}

func (u GeoFenceSettings) Validate() []*utils.ValidationError {

	return utils.GetValidate().Validate(u)
}

type VasFindRequest struct {
	IMEI string `json:"imei"`
	VAS  string `json:"vas"`
}

func (u VasFindRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (o *VasFindRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, o)
}

type VasFindResponse struct {
	utils.Result
	Subscriptions []*VasSubscription `json:"subscriptions,omitempty"`
	Vas           *Vas               `json:"vas,omitempty"`
}

func (u *VasFindResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type VasUpdateManySubscriptions struct {
	VAS       string      `json:"vas" validate:"nonzero"`
	IMEI      []string    `json:"imei" validate:"nonzero"`
	ProfileId string      `json:"profile_id" bson:"-" validate:"nonzero"`
	Settings  interface{} `json:"settings" bson:"settings"`
}

func (o *VasUpdateManySubscriptions) FromJSON(data []byte) error {

	err := json.Unmarshal(data, o)
	if err != nil {
		return err
	}

	return nil
}

type CloneSubsciptions struct {
	FromProfileId string `json:"from_profile_id" `
	IMEI          string `json:"imei"`
	ToProfileId   string `json:"to_profile_id"`
}

func (o *CloneSubsciptions) FromJSON(data []byte) error {

	err := json.Unmarshal(data, o)
	if err != nil {
		return err
	}

	return nil
}

func (u CloneSubsciptions) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

type UpdateSubscriptionStatus struct {
	ID     string `json:"id"` // this is id of subscription
	Status bool   `json:"status"`
}

func (o *UpdateSubscriptionStatus) FromJSON(data []byte) error {

	err := json.Unmarshal(data, o)
	if err != nil {
		return err
	}

	return nil
}

func (u UpdateSubscriptionStatus) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}
