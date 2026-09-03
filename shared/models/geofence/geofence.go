package geofence

import (
	"encoding/json"

	utils "blueassetgroup.com/reports-service/shared"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GeometryType string

const (
	GeometryLineString GeometryType = "LineString"
	GeometryPolygon    GeometryType = "Polygon"
)

type Line struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type Polygon struct {
	Type        string        `json:"type" validate:"nonzero"`
	Coordinates [][][]float64 `json:"coordinates" validate:"nonzero"`
}

type Geofence struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Profile     primitive.ObjectID `json:"profile_id"`
	Fence       any                `json:"fence"`
	//Devices     []string           `json:"devices"`
}

type GeofenceAddRequest struct {
	Name        string             `json:"name" validate:"nonzero"`
	Description string             `json:"description" validate:"nonzero"`
	Fence       Polygon            `json:"fence" validate:"nonzero"`
	Profile     primitive.ObjectID `json:"profile_id" validate:"nonzero"`
}

func (p *GeofenceAddRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p GeofenceAddRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

type GeofenceUpdateRequest struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" validate:"nonzero"`
	Description string             `json:"description" validate:"nonzero"`
}

func (p *GeofenceUpdateRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type GeofenceAddResponse struct {
	utils.Result
	ID string `json:"id"`
}

type GeoFenceAddDeviceRequest struct {
	FenceId string   `json:"id" validate:"nonzero"`
	Devices []string `json:"devices" validate:"nonzero"`
}

func (p *GeoFenceAddDeviceRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p GeoFenceAddDeviceRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

type GeoFenceAddDeviceResponse struct {
	utils.Result
}

type GeoFenceFindRequest struct {
	ProfileId  *string `json:"profile_id" validate:"nonzero"`
	GeofenceId *string `json:"geofence_id"`
	Detail     bool    `json:"detail"`
}

func (u GeoFenceFindRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (p *GeoFenceFindRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (p GeoFenceFindRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

type GeoFenceFindResponse struct {
	utils.Result
	Fences []*Geofence `json:"fences"`
	Fence  *Geofence   `json:"fence"`
}

func (u *GeoFenceFindResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type GeofenceRemoveRequest struct {
	ID string `json:"id"`
}

func (p *GeofenceRemoveRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type IsInsideRequest struct {
	FenceId string  `json:"fence_ud"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

func (p *IsInsideRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type IsInsideResponse struct {
	utils.Result
	FenceId string `json:"fence_id"`
	OK      bool   `json:"ok"`
}

func (p *IsInsideResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

func (u IsInsideRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}
