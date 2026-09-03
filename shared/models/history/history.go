package history

import (
	"encoding/json"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"blueassetgroup.com/reports-service/shared/models/realtime"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type HistoryTrip struct {
	Start           utils.RealtimeTime `json:"start"`
	End             utils.RealtimeTime `json:"end"`
	Count           int                `json:"count"`
	Imei            string             `json:"imei"`
	Distance        float64            `json:"distance"`
	MaxSpeed        float64            `json:"max_speed"`
	AvgSpeed        float64            `json:"average_speed"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	OdoStart        float32            `json:"odo_start"`
	OdoEnd          float32            `json:"odo_end"`
	Business        bool               `json:"business"`
	FuelCost        float32            `json:"fuel_cost"`
	MaintenanceCost float32            `json:"maintenance_cost"`
	Points          []*LatLng          `json:"points,omitempty"`
	StartAddress    string             `json:"start_address"`
	EndAddress      string             `json:"end_address"`
	WorkingTime     float64            `json:"working_time"`
	Driver          string             `json:"driver"`
}

type LatLng struct {
	ID        primitive.ObjectID `json:"id" bson:"id"`
	Lat       float64            `json:"lat"`
	Lon       float64            `json:"lon"`
	Direction string             `json:"direction"`
	Speed     float64            `json:"speed"`
	Alt       float64            `json:"alt"`
	Address   string             `json:"address"`
	Timestamp time.Time          `json:"timestamp"`
}

const (
	AddressToDisplayStreet  = "street"
	AddressToDisplayMyplace = "myplace"
	AddressToDisplayBoth    = "both"
)

type HistoryTripRequest struct {
	Imei             string           `json:"imei" validate:"nonzero"`
	Start            utils.CustomTime `json:"start" validate:"nonzero"`
	End              utils.CustomTime `json:"end" validate:"nonzero"`
	ExludePoints     any              `json:"exclude_points"`
	DriverInfo       any              `json:"driver_info"`
	AddressToDisplay string           `json:"address_to_display"`
	ProfileId        string           `json:"profile_id"`
}

func (u *HistoryTripRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u HistoryTripRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u HistoryTripRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

type HistoryTripResponse struct {
	utils.Result
	Imei  string         `json:"imei"`
	Start string         `json:"start"`
	End   string         `json:"end"`
	Trips []*HistoryTrip `json:"trips"`
}

func (u *HistoryTripResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type HistoryPointRequest struct {
	ID   string `json:"id"`
	Imei string `json:"imei"`
}

func (u *HistoryPointRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type HistoryPointResponse struct {
	utils.Result
	Point *realtime.Message `json:"point"`
}

type HistoryTripSaveRequest struct {
	Trips []struct {
		Imei            string             `json:"imei" validate:"nonzero"`
		Start           utils.RealtimeTime `json:"start" validate:"nonzero"`
		End             utils.RealtimeTime `json:"end" validate:"nonzero"`
		Business        bool               `json:"business"`
		FuelCost        float32            `json:"fuel_cost"`
		MaintenanceCost float32            `json:"maintenance_cost"`
		Description     string             `json:"description"`
	} `json:"trips"`
}

func (u *HistoryTripSaveRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u HistoryTripSaveRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

type Trip struct {
	ID              string  `json:"id" bson:"_id"`
	Imei            string  `json:"imei"`
	Business        bool    `json:"business"`
	FuelCost        float32 `json:"fuel_cost"`
	MaintenanceCost float32 `json:"maintenance_cost"`
	Description     string  `json:"description"`
}
