package profiles

import (
	"encoding/json"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContactGroup struct {
	ID          string             `json:"id" bson:"-"`
	MongoID     primitive.ObjectID `json:"-" bson:"_id"`
	ProfileId   string             `json:"profile_id" bson:"profile_id"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Contacts    []string           `json:"contacts" bson:"contacts"`
	LastUpdated time.Time          `json:"last_updated" bson:"last_updated"`
}

func (_self *ContactGroup) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type DeleteContactRequest struct {
	ContactId string `json:"contact_id"`
}

func (_self *DeleteContactRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type GetContactRequest struct {
	ProfileId string `json:"profile_id"`
}

func (_self *GetContactRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type AddContactsToGroupRequest struct {
	ProfileId string   `json:"profile_id"`
	GroupId   string   `json:"group_id"`
	Contacts  []string `json:"contacts" bson:"contacts"`
}

func (_self *AddContactsToGroupRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type DeleteContactsToGroupRequest struct {
	ProfileId string   `json:"profile_id"`
	GroupId   string   `json:"group_id"`
	Contacts  []string `json:"contacts" bson:"contacts"`
}

func (_self *DeleteContactsToGroupRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type AddGroupsToContactRequest struct {
	Groups []string `json:"groups" bson:"groups"`
}

func (_self *AddGroupsToContactRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type License struct {
	Codes          string `json:"codes" bson:"codes"`
	Expiry         string `json:"expiry,omitempty" bson:"expiry,omitempty"`
	ExpiryReminder string `json:"expiry_reminder" bson:"expiry_reminder"`
}

type Contact struct {
	ID           string             `json:"id" bson:"-" `
	MongoID      primitive.ObjectID `json:"-" bson:"_id"`
	ProfileId    string             `json:"profile_id" bson:"profile_id" validate:"nonzero"`
	Name         string             `json:"name,omitempty" bson:"name,omitempty" validate:"nonzero"`
	Surname      string             `json:"surname,omitempty" bson:"surname,omitempty" validate:"nonzero"`
	Email        string             `json:"email,omitempty" bson:"email,omitempty" validate:"nonzero"`
	ContactNo    string             `json:"contact_no,omitempty" bson:"contact_no,omitempty"`
	DriverKey    string             `json:"driver_key,omitempty" bson:"driver_key,omitempty"`
	License      *License           `json:"license,omitempty" bson:"license,omitempty"`
	LastUpdated  time.Time          `json:"last_updated" bson:"last_updated"`
	Relationship string             `json:"relationship,omitempty" bson:"relationship,omitempty"`
}

func (_self *Contact) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

func (u Contact) ToJSON() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (p Contact) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

type FindAllContactsRequest struct {
	ProfileId string           `json:"profile_id"`
	Filter    utils.FindFilter `json:"filter"`
}

func (_self *FindAllContactsRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

func (u FindAllContactsRequest) ToJSON() []byte {
	b, _ := json.Marshal(u)
	return b
}

type FindAllContactsResult struct {
	utils.Result
	ProfileId string     `json:"profile_id"`
	Contacts  []*Contact `json:"contacts"`
	Count     int64      `json:"total_count"`
}

func (_self *FindAllContactsResult) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type FindContactRequest struct {
	ProfileId string `json:"profile_id"`
	ContactId string `json:"contact_id"`
	DriverKey string `json:"driver_key"`
}

func (_self *FindContactRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

func (u FindContactRequest) ToJSON() ([]byte, error) {
	return json.Marshal(u)
}

type FindContactResponse struct {
	utils.Result
	Contact *Contact `json:"contact"`
}

func (_self *FindContactResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type FindAllContactGroupsResult struct {
	utils.Result
	ProfileId string          `json:"profile_id"`
	Groups    []*ContactGroup `json:"groups"`
}

type FindAllContactGroupsRequest struct {
	ProfileId string `json:"profile_id"`
}

func (_self *FindAllContactGroupsRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type DeleteContactGroupRequest struct {
	GroupId   string `json:"group_id"`
	ProfileId string `json:"profile_id"`
}

func (_self *DeleteContactGroupRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}

type GetContactGroupRequest struct {
	ProfileId string `json:"profile_id"`
	GroupId   string `json:"group_id"`
}

type GetContactGroupResponse struct {
	utils.Result
	ContactGroup *ContactGroup `json:"contact_group"`
}

func (_self *GetContactGroupRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, _self)
}
