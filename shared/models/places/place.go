package places

import (
	"encoding/json"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// { "_id" : ObjectId("571f5b7e60b22bac370be3f5"),
// "userId" : "39ffd710-0705-11e6-8377-001e67aa4714",
// "name" : "Bryns House",
// "lat" : -25.74842690925808,
// "lon" : 28.325344920158386,
// "radius" : 5,
/// "iconColor" : "blue",
//"address" : "304 Rotsvygie St, Pretoria, 0184, South Africa" }

type Point struct {
	Type string `json:"type" bson:"type"`
	// lon, lat
	Coordinates []float64 `json:"coordinates" bson:"coordinates"`
}
type Place struct {
	ID             string             `json:"place_id" bson:"-" `
	MongoID        primitive.ObjectID `json:"-" bson:"_id"`
	ProfileId      string             `json:"profile_id" bson:"-" validate:"nonzero"`
	MongoProfileId primitive.ObjectID `json:"-" bson:"profile_id"`
	Name           string             `json:"name,omitempty" bson:"name,omitempty" validate:"nonzero"`
	Email          string             `json:"email,omitempty" bson:"email,omitempty" `
	Person         string             `json:"person,omitempty" bson:"person,omitempty" `
	Mobile         string             `json:"mobile,omitempty" bson:"mobile,omitempty" `
	Tel            string             `json:"tel,omitempty" bson:"tel,omitempty" `
	Web            string             `json:"web,omitempty" bson:"web,omitempty" `
	Address        string             `json:"address,omitempty" bson:"address,omitempty" `
	LastUpdated    time.Time          `json:"last_updated" bson:"last_updated"`
	Lat            float64            `json:"lat" bson:"-" validate:"nonzero"`
	Lon            float64            `json:"lon" bson:"-" validate:"nonzero"`
	Radius         int16              `json:"radius" bson:"radius" validate:"nonzero"`
	IconColour     string             `json:"colour" bson:"colour"`
	Point          *Point             `json:"-" bson:"point"`
}

func (p Place) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(p)
}

func (p *Place) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type FindPlacesRequest struct {
	ProfileId string `json:"profile_id"`
	PlaceId   string `json:"place_id"`
}

func (u FindPlacesRequest) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}

func (p *FindPlacesRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type FindPlacesResult struct {
	utils.Result
	ProfileId string   `json:"profile_id"`
	Places    []*Place `json:"places"`
	Count     int      `json:"total_count"`
}

func (p *FindPlacesResult) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type FindPlaceResult struct {
	utils.Result
	Place *Place `json:"place"`
}

type AddPlaceResult struct {
	utils.Result
	ID string `json:"id" bson:"-" `
}

type DeletePlacesRequest struct {
	PlaceId string   `json:"place_id"`
	Places  []string `json:"places"`
}

func (p *DeletePlacesRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

type PlaceGroup struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	ProfileId   string             `json:"profile_id" bson:"profile_id"`
	Name        string             `json:"name,omitempty"`
	Description string             `json:"description,omitempty"`
	Places      []string           `json:"places,omitempty" bson:"places"`
}
