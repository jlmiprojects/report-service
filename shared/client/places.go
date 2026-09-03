package client

import (
	"context"
	"errors"
	"time"

	places "blueassetgroup.com/reports-service/shared/models/places"
	"github.com/nats-io/nats.go"
)

func FindAllPlaces(nc *nats.Conn, profileId string) ([]*places.Place, error) {

	request := places.FindPlacesRequest{}

	request.ProfileId = profileId

	j := request.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, "places.findall", j)

	if err != nil {
		return nil, err
	}

	response := new(places.FindPlacesResult)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 && response.StatusCode != 0 {
		return nil, errors.New(response.Message)
	}

	return response.Places, nil

}
