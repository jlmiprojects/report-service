package client

import (
	"context"
	"errors"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"blueassetgroup.com/reports-service/shared/models/profiles"
	profile "blueassetgroup.com/reports-service/shared/models/profiles"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddProfile(nc *nats.Conn, profile *profiles.Profile) (string, error) {

	j, err := profile.ToJSON()

	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILES_ADD, j)

	if err != nil {
		return "", err
	}

	response := new(profiles.CreateProfileResult)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return "", err
	}

	if response.StatusCode != 200 {
		return "", errors.New(response.Error)
	}

	return response.ID, nil

}

func GetContactById(nc *nats.Conn, id string) (*profile.Contact, error) {

	req := profile.FindContactRequest{
		ContactId: id,
	}

	j, err := req.ToJSON()

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILES_CONTACTS_FIND, j)

	if err != nil {
		return nil, err
	}

	response := new(profile.FindContactResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Contact, nil

}

func GetProfileById(nc *nats.Conn, id string) (*profile.Profile, error) {

	req := profile.FindProfileRequest{
		ID: id,
	}

	j, err := req.ToJSON()

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILES_FIND, j)

	if err != nil {
		return nil, err
	}

	response := new(profile.FindProfileResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Profile, nil

}

func GetProfileByUserName(nc *nats.Conn, username string) (*profile.Profile, error) {

	req := profile.FindProfileRequest{
		UserName: username,
	}

	j, err := req.ToJSON()

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILES_FIND, j)

	if err != nil {
		return nil, err
	}

	response := new(profile.FindProfileResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Profile, nil

}

func GetAllProfileDevices(nc *nats.Conn, profileID string) ([]*profile.ProfileDevice, error) {

	req := profile.ProfileDeviceFindAllRequest{
		ProfileId: profileID,
	}

	j, err := req.ToJSON()

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILE_DEVICES_FIND_ALL, j)

	if err != nil {
		return nil, err
	}

	response := new(profile.ProfileDeviceFindAllResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Data, nil

}

func FindAllProfileDevicesByGroupId(nc *nats.Conn, profileID string, groupId string) ([]string, error) {

	req := profile.ProfileDeviceGroupFindRequest{
		ProfileId:   profileID,
		GroupId:     groupId,
		ShowDetails: false,
	}

	j, err := req.ToJSON()

	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILES_DEVICES_GROUPS_FIND, j)

	if err != nil {
		return nil, err
	}

	response := new(profile.ProfileDeviceGroupFindResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	// This is wierd as there should only be one
	if len(response.Groups) > 0 {

		if _temp, ok := response.Groups[0].Devices.([]interface{}); ok {

			_devices, ok := interfaceSliceToStringSlice(_temp)

			if !ok {
				return nil, errors.New("failed to convert devices to string slice")
			}
			return _devices, nil
		} else {
			return make([]string, 0), nil
		}

	} else {
		return make([]string, 0), nil
	}

}

func FindAllProfiles(nc *nats.Conn, params string) ([]*profile.Profile, int64, error) {

	req := profile.FindAllProfileRequest{
		Params: params,
	}

	j, err := req.ToJSON()

	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILES_FINDALL, j)

	if err != nil {
		return nil, 0, err
	}

	response := new(profile.FindProfilesResult)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, 0, err
	}

	/*if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}*/

	return response.Profiles, response.Count, nil

}

func AddDeviceToProfile(nc *nats.Conn, profileId string, imei string) error {

	_profileId, err := primitive.ObjectIDFromHex(profileId)

	if err != nil {
		return err
	}

	req := profile.ProfileDevice{
		ProfileId: _profileId,
		IMEI:      imei,
	}

	j, err := req.ToJSON()

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILE_DEVICES_ADD, j)

	if err != nil {
		return err
	}

	response := new(utils.Result)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return err
	}

	return nil

}

func SubProfileAdd(nc *nats.Conn, parent string, children []string) error {

	req := profile.SubProfileAddRequest{}
	req.Parent = parent
	req.Children = children

	j, err := req.ToJSON()

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILE_SUB_PROFILE_ADD, j)

	if err != nil {
		return err
	}

	response := new(utils.Result)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return err
	}

	return nil

}

func interfaceSliceToStringSlice(interfaceSlice []interface{}) ([]string, bool) {
	stringSlice := make([]string, len(interfaceSlice))
	for i, element := range interfaceSlice {
		str, ok := element.(string)
		if !ok {
			return nil, false // Return nil slice and false if any element is not a string
		}
		stringSlice[i] = str
	}
	return stringSlice, true
}
