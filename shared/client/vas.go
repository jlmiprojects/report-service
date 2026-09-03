package client

import (
	"context"
	"errors"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	model "blueassetgroup.com/reports-service/shared/models/vas"
	"github.com/nats-io/nats.go"
)

func FindSubscriptionsByImeiAndVas(nc *nats.Conn, imei string, vas string) ([]*model.VasSubscription, error) {
	req := model.VasFindRequest{
		IMEI: imei,
		VAS:  vas,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.VAS_FIND_SUBSCRIPTION, j)

	if err != nil {
		return nil, err
	}

	response := new(model.VasFindResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Subscriptions, nil
}

func FindSubscriptionsByImeiAndProfileId(nc *nats.Conn, imei string, profileId string) ([]string, error) {

	req := model.VasFindAllSubscriptionsRequest{
		IMEI:      imei,
		ProfileId: profileId,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.VAS_FIND_SUBSCRIPTIONS, j)

	if err != nil {
		return nil, err
	}

	response := new(model.VasFindAllSubscriptionsResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	retval := make([]string, len(response.Data))

	for i, vas := range response.Data {
		retval[i] = vas.VAS
	}

	return retval, nil
}

func CloneSubsciptions(nc *nats.Conn, from string, imei string, to string) error {

	req := model.CloneSubsciptions{
		IMEI:          imei,
		FromProfileId: from,
		ToProfileId:   to,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.VAS_CLONE_SUBSCRIPTIONS, j)

	if err != nil {
		return err
	}

	response := new(utils.Result)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return errors.New(response.Error)
	}

	return nil
}

func Subscribe(nc *nats.Conn, imei string, profile_id string, subscriptions []string) error {

	req := model.VasSubscriptionsRequest{
		IMEI:          imei,
		ProfileId:     profile_id,
		Subscriptions: subscriptions,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.VAS_SUBSCRIBE, j)

	if err != nil {
		return err
	}

	response := new(utils.Result)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return err
	}

	if response.StatusCode != 201 {
		return errors.New(response.Message)
	}

	return nil

}

func UpdateSubscriptionStatus(nc *nats.Conn, id string, status bool) error {

	req := model.UpdateSubscriptionStatus{
		ID:     id,
		Status: status,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.VAS_UPDATE_SUBSCRIPTION_STATUS, j)

	if err != nil {
		return err
	}

	response := new(utils.Result)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return errors.New(response.Message)
	}

	return nil

}
