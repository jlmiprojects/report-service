package mongo

import (
	"encoding/json"

	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson"
)

type MongoRequest struct {
	TTL        string `json:"ttl"`
	Database   string `json:"database"`
	Collection string `json:"collection"`
}

type MongoCountRequest struct {
	MongoRequest
	Filter bson.M `json:"filter"`
}

func (u MongoCountRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *MongoCountRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type MongoCountResponse struct {
	utils.Result
	Count int64 `json:"count"`
}

func (u *MongoCountResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type MongoFindRequest struct {
	MongoRequest
	Filter     bson.M `json:"filter,omitempty"`
	Projection bson.M `json:"projection,omitempty"`
	Limit      int64  `json:"limit"`
	Skip       int64  `json:"skip"`
	Sort       bson.M `json:"sort"`
}

func (u MongoFindRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *MongoFindRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type MongoFindResponse struct {
	utils.Result
	Data  []map[string]any `json:"data"`
	Count int              `json:"count"`
}

func (u *MongoFindResponse) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type MongoAggregateRequest struct {
	MongoRequest
	Pipeline []bson.M `json:"pipeline"`
}

func (u MongoAggregateRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (u *MongoAggregateRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}
