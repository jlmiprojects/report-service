package client

import (
	"context"
	"errors"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	model "blueassetgroup.com/reports-service/shared/models/alert"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FindAlerts(nc *nats.Conn, ff utils.FindFilter) ([]*model.Alert, int64, error) {

	j := ff.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.ALERT_FIND_ALL, j)

	if err != nil {
		return nil, 0, err
	}

	response := new(model.AlertFindAllResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode != 200 {
		return nil, 0, errors.New(response.Error)
	}

	return response.Data, response.Count, nil

}

func FindAlertById(nc *nats.Conn, alertId string) (*model.Alert, error) {

	req := model.AlertFindRequest{AlertId: alertId}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.ALERT_FIND, j)

	if err != nil {
		return nil, err
	}

	response := new(model.AlertFindResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Alert, nil

}

func AddEvent(nc *nats.Conn, alertId string, comment string, userName string, status string) error {

	req := model.AddAlertEventRequest{
		AlertId:  alertId,
		Comment:  comment,
		UserName: userName,
		Status:   status,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.ALERT_ADD_EVENT, j)

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

// AddCallIn records a call-in as an Alert with AlertType "CALLIN", using the
// operator-supplied comment as the initial event instead of the generic
// "Alert Created" text AddAlert uses for system-generated alerts. Returns the
// new alert's id.
func AddCallIn(nc *nats.Conn, imei string, profileid string, comment string, userName string) (string, error) {

	_profile, err := primitive.ObjectIDFromHex(profileid)

	if err != nil {
		return "", err
	}

	req := model.Alert{
		IMEI:      imei,
		Profile:   _profile,
		Timestamp: time.Now(),
		AlertType: "CALLIN",
		Events:    make([]model.AlertEvent, 1),
	}

	req.Events[0] = model.AlertEvent{
		Timestamp: time.Now(),
		Comment:   comment,
		Status:    model.ALERT_STATUS_CREATED.String(),
		UserName:  userName,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.ALERT_ADD, j)

	if err != nil {
		return "", err
	}

	response := new(model.AlertAddResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", errors.New(response.Error)
	}

	return response.AlertId, nil
}

func AddAlert(nc *nats.Conn, imei string, profileid string, alertType string) error {

	_profile, err := primitive.ObjectIDFromHex(profileid)

	if err != nil {
		return err
	}
	req := model.Alert{
		IMEI:      imei,
		Profile:   _profile,
		Timestamp: time.Now(),
		AlertType: alertType,
		Events:    make([]model.AlertEvent, 1),
	}

	req.Events[0] = model.AlertEvent{
		Timestamp: time.Now(),
		Comment:   "Alert Created",
		Status:    model.ALERT_STATUS_CREATED.String(),
		UserName:  "System",
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.ALERT_ADD, j)

	if err != nil {
		return err
	}

	response := new(model.AlertAddResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return errors.New(response.Error)
	}

	return nil
}
