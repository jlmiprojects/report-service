package client

import (
	"context"
	"errors"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"blueassetgroup.com/reports-service/shared/models/devices"
	device "blueassetgroup.com/reports-service/shared/models/devices"
	"github.com/nats-io/nats.go"
)

func AssignOwner(nc *nats.Conn, packet *devices.AssignOwnerRequest) error {

	j := packet.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.DEVICES_UPDATE_OWNER, j)

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

func ProvisionDevice(nc *nats.Conn, device *devices.ProvisionRequest) (string, error) {

	j := device.ToJSON()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.DEVICES_PROVISION, j)

	if err != nil {
		return "", err
	}

	response := new(devices.ProvisionResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		if len(response.Error) > 0 {
			return "", errors.New(response.Error)
		} else {
			return "", errors.New(response.Message)
		}
	}

	return response.Imei, nil

}

func FindDeviceByIMEI(nc *nats.Conn, imei string, ttl time.Duration) (*device.Device, error) {

	req := device.FindDeviceRequest{
		IMEI: imei,
	}

	msg := nats.NewMsg(utils.DEVICES_FINDONE)
	msg.Data = req.ToBytes()
	msg.Header.Add("ttl", ttl.String())
	msg, err := nc.RequestMsg(msg, ttl)

	if err != nil {
		return nil, err
	}

	response := new(device.FindDeviceResult)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Device, nil

}

func FindDevicesByProfileId(nc *nats.Conn, profile_id string) ([]*device.Device, error) {

	req := utils.FindFilter{
		ProfileId: profile_id,
		Owner:     profile_id,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.DEVICES_FINDALL, j)

	if err != nil {
		return nil, err
	}

	response := new(device.FindDevicesResult)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Devices, nil

}

func FindAll(nc *nats.Conn, params string) ([]*device.Device, error) {

	req := utils.FindFilter{
		Params: params,
	}

	j := req.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*180)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.DEVICES_FINDALL, j)

	if err != nil {
		return nil, err
	}

	response := new(device.FindDevicesResult)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Devices, nil

}
