package client

import (
	"context"
	"errors"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"blueassetgroup.com/reports-service/shared/models/geofence"
	"github.com/nats-io/nats.go"
)

func FindGeofence(nc *nats.Conn, fenceId string) (*geofence.Geofence, error) {

	req := geofence.GeoFenceFindRequest{
		GeofenceId: &fenceId,
		Detail:     false,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.GEOFENCE_FIND, j)

	if err != nil {
		return nil, err
	}

	response := new(geofence.GeoFenceFindResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Fence, nil
}

func IsInside(nc *nats.Conn, fenceId string, lat float64, lon float64) (bool, error) {

	req := geofence.IsInsideRequest{
		FenceId: fenceId,
		Lat:     lat,
		Lon:     lon,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.GEOFENCE_INSIDE, j)

	if err != nil {
		return false, err
	}

	response := new(geofence.IsInsideResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return false, err
	}

	if response.StatusCode != 200 {
		return false, errors.New(response.Error)
	}

	return response.OK, nil
}
