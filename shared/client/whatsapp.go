package client

import (
	"context"
	"errors"
	"time"

	utils "blueassetgroup.com/reports-service/shared"
	sms "blueassetgroup.com/reports-service/shared/models/sms"
	"github.com/nats-io/nats.go"
)

func SendWhatsApp(nc *nats.Conn, to string, template string, data any, async bool) error {

	req := new(sms.WASendRequest)

	req.To = to //string{to}
	req.Template = template
	req.Data = data

	b := req.ToBytes()

	if async {
		err := nc.Publish(utils.WHATSAPP_SEND, b)
		if err != nil {
			return err
		}
		return nil

	} else {

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		defer cancel()

		msg, err := nc.RequestWithContext(ctx, utils.WHATSAPP_SEND, b)

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
