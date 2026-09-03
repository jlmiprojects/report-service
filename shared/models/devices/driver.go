package devices

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Driver struct {
	ID             string             `json:"id" bson:"-"`
	MongoID        primitive.ObjectID `bson:"_id" json:"-"`
	Name           string             `json:"name" bson:"name"`
	Surname        string             `json:"surname" bson:"surname"`
	Email          string             `json:"email" bson:"email"`
	ContactNumber  string             `json:"contact_number" bson:"contact_number"`
	DriverKey      string             `json:"driver_key" bson:"driver_key"`
	LicenseCode    string             `json:"license_code" bson:"license_code"`
	LicenseRenewal *time.Time         `json:"license_renewal" bson:"license_renewal"`
}
