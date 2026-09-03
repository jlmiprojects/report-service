package client

import (
	"context"
	"errors"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	email "blueassetgroup.com/reports-service/shared/models/email"
	"github.com/nats-io/nats.go"
)

func SendEmail(nc *nats.Conn, from string, to string, subject string, template string, data any, async bool) error {

	req := new(email.Email)

	req.To = []string{to}
	req.Template = template
	req.Data = data
	req.From = from
	req.Subject = subject

	b := req.ToBytes()

	if async {
		err := nc.PublishRequest(utils.EMAIL_SEND, "email_response", b)
		if err != nil {
			return err
		}
		return nil

	} else {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		defer cancel()

		msg, err := nc.RequestWithContext(ctx, utils.EMAIL_SEND, b)

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

	}

	return nil

}
