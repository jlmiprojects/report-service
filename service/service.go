package service

import (
	"context"
	"encoding/json"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"blueassetgroup.com/reports-service/shared/client"
	"blueassetgroup.com/reports-service/shared/models/devices"
	"blueassetgroup.com/reports-service/shared/models/history"
	"blueassetgroup.com/reports-service/shared/models/places"
	"blueassetgroup.com/reports-service/shared/models/profiles"
	"github.com/nats-io/nats.go"
)

type ServiceCall struct {
	nc *nats.Conn
}

func NewServiceCall(nc *nats.Conn) (*ServiceCall, error) {
	return &ServiceCall{nc: nc}, nil
}

func (service ServiceCall) FindAllContacts(profileId string) ([]*profiles.Contact, error) {

	return client.FindAllContacts(service.nc, profileId)
}

func (service ServiceCall) FindAllDevicesByProfileId(profileId string) ([]*devices.Device, error) {

	return client.FindDevicesByProfileId(service.nc, profileId)
}

func (service ServiceCall) FindProfile(profileId string) (*profiles.Profile, error) {

	return client.GetProfileById(service.nc, profileId)
}

func (service ServiceCall) FindAllTrips(profileId string, imei string, start utils.CustomTime, end utils.CustomTime, addressToDisplay string) ([]*history.HistoryTrip, error) {
	return client.FindTripsByImeiAndStartandEndandAddressToDisplay(service.nc, profileId, imei, start, end, true, addressToDisplay)
}

func (service ServiceCall) FindAllPlaces(profileId string) ([]*places.Place, error) {
	return client.FindAllPlaces(service.nc, profileId)
}

func (service ServiceCall) Generic(action string, ttl time.Duration, request map[string]any) (any, error) {

	ctx, cancel := context.WithTimeout(context.Background(), ttl)

	defer cancel()

	b, _ := json.Marshal(request)

	msg, err := service.nc.RequestWithContext(ctx, action, b)

	if err != nil {
		return nil, err
	}

	response := make(map[string]any)
	err = json.Unmarshal(msg.Data, &response)

	if err != nil {
		return nil, err
	}

	return response, nil

}

func (service ServiceCall) FindDevice(imei string) (*devices.Device, error) {

	return client.FindDeviceByIMEI(service.nc, imei, 10*time.Second)
}
