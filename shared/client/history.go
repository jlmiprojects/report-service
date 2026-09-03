package client

import (
	"context"
	"errors"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	history "blueassetgroup.com/reports-service/shared/models/history"
	"blueassetgroup.com/reports-service/shared/models/realtime"
	"github.com/nats-io/nats.go"
)

func LastPoint(nc *nats.Conn, imeis []string) ([]*realtime.Message, error) {

	request := history.LastRequest{AlwaysArray: true, Imeis: imeis}

	j := request.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.HISTORY_LAST, j)

	if err != nil {
		return nil, err
	}

	response := new(history.LastResponse2)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	if response.Messages == nil {
		return make([]*realtime.Message, 0), nil
	}

	return response.Messages, nil

}

func FindTripsByImeiAndStartandEnd(nc *nats.Conn, imei string, start utils.CustomTime, end utils.CustomTime, excludePoints bool) ([]*history.HistoryTrip, error) {

	request := history.HistoryTripRequest{}

	request.Imei = imei
	request.Start = start
	request.End = end
	request.ExludePoints = excludePoints

	j := request.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.HISTORY_TRIPS, j)

	if err != nil {
		return nil, err
	}

	response := new(history.HistoryTripResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Trips, nil

}

func FindTripsByImeiAndStartandEndandAddressToDisplay(nc *nats.Conn, profileId string, imei string, start utils.CustomTime, end utils.CustomTime, excludePoints bool, addressToDisplay string) ([]*history.HistoryTrip, error) {

	request := history.HistoryTripRequest{}

	request.Imei = imei
	request.Start = start
	request.End = end
	request.ExludePoints = excludePoints
	request.AddressToDisplay = addressToDisplay
	request.ProfileId = profileId

	j := request.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.HISTORY_TRIPS, j)

	if err != nil {
		return nil, err
	}

	response := new(history.HistoryTripResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Trips, nil

}
