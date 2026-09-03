package realtime

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	Battery         = "bat"
	InternalBattery = "ibat"
	Movement        = "movement"
	Tamper          = "tamper"
	Speed           = "speed"
	Imei            = "imei"
	Lat             = "lat"
	Lon             = "lon"
	GpsLock         = "gpslock"
	Heading         = "bearing"
	Ignition        = "ignition"
	Panic           = "panic"
	Satelites       = "satelites"
	Altitude        = "alt"
	Timestamp       = "timestamp"
	EventTimestamp  = "event_timestamp"
	Odo             = "odo"
	Analog1         = "analog1"
	Analog2         = "analog2"
	iButton         = "driverkey"
	HarshAccel      = "harsh_accel"
	HarshBrake      = "harsh_brake"
	HarshCorner     = "harsh_corner"
	HarshValue      = "harsh_value"
	Digital1        = "digital1"
	Digital2        = "digital2"
	Fuel1           = "fuel1"
	Fuel2           = "fuel2"
	Fuel3           = "fuel3"
	Fuel4           = "fuel4"
	BLEBattery1     = "ble_bat1"
	BLEBattery2     = "ble_bat2"
	BLEBattery3     = "ble_bat3"
	BLEBattery4     = "ble_bat4"
)

type Message struct {
	ID                         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Imei                       string             `json:"imei" bson:"imei"`
	Timestamp                  string             `json:"timestamp" bson:"-"`
	EventTimestamp             time.Time          `json:"event_timestamp" bson:"event_timestamp"`
	MongoTimestamp             time.Time          `bson:"timestamp" json:"-"`
	DeviceType                 string             `json:"device_type"`
	Lat                        float64            `json:"lat,omitempty"`
	Lon                        float64            `json:"lon,omitempty"`
	Speed                      float64            `json:"speed"`
	GpsLock                    bool               `json:"gpslock"`
	Alt                        float64            `json:"alt,omitempty"`
	Heading                    float32            `json:"bearing,omitempty"`
	Ignition                   string             `json:"ignition,omitempty"`
	Panic                      bool               `json:"sos"`
	Crash                      bool               `json:"crash"`
	Tamper                     bool               `json:"tamper"`
	Movement                   bool               `json:"movement"`
	Satelites                  int16              `json:"satelites"`
	ExternalBattery            float32            `json:"xbatt,omitempty"`
	BatteryPercentage          int                `json:"batt_perc,omitempty"`
	BatteryVoltage             float32            `json:"batt_volt,omitempty"`
	Odo                        float32            `json:"odo,omitempty"`
	Analog1                    float32            `json:"analog1"`
	Analog2                    float32            `json:"analog2"`
	Digital1                   int                `json:"digital_in1"`
	Digital2                   int                `json:"digital_in2"`
	DigitalOut1                int                `json:"digital_out1"`
	DigitalOut2                int                `json:"digital_out2"`
	IButton                    string             `json:"driverkey,omitempty"`
	HarshCorner                bool               `json:"harsh_corner,omitempty"`
	HarshBrake                 bool               `json:"harsh_brake,omitempty"`
	HarshAccel                 bool               `json:"harsh_accel,omitempty"`
	HarshValue                 float64            `json:"harsh_value,omitempty"`
	Direction                  string             `json:"direction,omitempty"`
	VIN                        string             `json:"vin,omitempty"`
	RPM                        uint16             `json:"rpm,omitempty"`
	Fuel1                      float32            `json:"fuel1,omitempty"`
	Fuel2                      float32            `json:"fuel2,omitempty"`
	Fuel3                      float32            `json:"fuel3,omitempty"`
	Fuel4                      float32            `json:"fuel4,omitempty"`
	ProximityViolationState    uint8              `json:"proximity_violation_state,omitempty"`
	ProximityViolationSource   string             `json:"proximity_violation_source,omitempty"`
	ProximityViolationDuration uint16             `json:"proximity_violation_duration,omitempty"`
	BLEBat1                    uint8              `json:"ble_bat1,omitempty"`
	BLEBat2                    uint8              `json:"ble_bat2,omitempty"`
	BLEBat3                    uint8              `json:"ble_bat3,omitempty"`
	BLEBat4                    uint8              `json:"ble_bat4,omitempty"`
	Temperature                int                `json:"temperature,omitempty"`
	Status                     string             `json:"status"`
	Jamming                    bool               `json:"jamming"`
	Towing                     bool               `json:"towing"`
	Idling                     bool               `json:"idling"`
	Address                    string             `json:"address"`
	SpeedLimit                 int                `json:"speed_limit"`
}

func (message *Message) FromJson(data []byte) error {

	return json.Unmarshal(data, message)
}

func (message *Message) ToJson() ([]byte, error) {
	return json.Marshal(message)
}
