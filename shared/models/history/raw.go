package history

import (
	"encoding/json"

	utils "blueassetgroup.com/reports-service/shared"
	"blueassetgroup.com/reports-service/shared/models/realtime"
)

type RawRequest struct {
	Start  utils.CustomTime `json:"start"  validate:"nonzero"`
	End    utils.CustomTime `json:"end"  validate:"nonzero"`
	IMEI   string           `json:"imei"  validate:"nonzero,len=15"`
	Fields []string         `json:"fields"`
	Params string           `json:"params"`
}

type RawResponse struct {
	utils.Result `json:"result"`
	Data         []*realtime.Message `json:"data"`
	Count        int                 `json:"count"`
}

func (u *RawRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u RawRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

type LastRequest struct {
	Imei        string   `json:"imei"`
	Imeis       []string `json:"data,omitempty"`
	AlwaysArray bool     `json:"alway_array"`
}

func (u *LastRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u LastRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

type LastResponse struct {
	utils.Result
	Messages any `json:"data,omitempty"`
}

func (u *LastResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u LastResponse) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

type LastResponse2 struct {
	utils.Result
	Messages []*realtime.Message `json:"data,omitempty"`
}

func (u *LastResponse2) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u LastResponse2) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}
