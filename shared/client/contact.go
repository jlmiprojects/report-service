package client

import (
	"context"
	"errors"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	"blueassetgroup.com/reports-service/shared/models/profiles"
	profile "blueassetgroup.com/reports-service/shared/models/profiles"
	"github.com/nats-io/nats.go"
)

func AddContact(nc *nats.Conn, owner string, name string, surname string, msisdn string, email string, relationship string) error {

	req := profiles.Contact{
		ProfileId:    owner,
		Name:         name,
		Surname:      surname,
		ContactNo:    msisdn,
		Email:        email,
		Relationship: relationship,
	}

	j := req.ToJSON()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILES_CONTACTS_ADD, j)

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

func FindContactById(nc *nats.Conn, contactId string) (*profile.Contact, error) {

	req := profiles.FindContactRequest{
		ContactId: contactId,
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

	response := new(profiles.FindContactResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Contact, nil

}

func FindContactByDriverKey(nc *nats.Conn, profileId string, driverKey string) (*profile.Contact, error) {

	req := profiles.FindContactRequest{
		ProfileId: profileId,
		DriverKey: driverKey,
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

	response := new(profiles.FindContactResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Contact, nil

}

func FindAllContacts(nc *nats.Conn, profileId string) ([]*profiles.Contact, error) {

	req := profiles.FindAllContactsRequest{
		ProfileId: profileId,
	}

	j := req.ToJSON()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, utils.PROFILES_CONTACTS_FINDALL, j)

	if err != nil {
		return nil, err
	}

	response := new(profiles.FindAllContactsResult)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(response.Error)
	}

	return response.Contacts, nil

}
