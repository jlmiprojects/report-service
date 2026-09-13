package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

type ServiceCall struct {
	nc *nats.Conn
}

func NewServiceCall(nc *nats.Conn) (*ServiceCall, error) {
	return &ServiceCall{nc: nc}, nil
}

func (service ServiceCall) Generic(action string, ttl time.Duration, request map[string]any) (any, error) {

	ctx, cancel := context.WithTimeout(context.Background(), ttl)

	defer cancel()

	b, _ := json.Marshal(request)

	msg, err := service.nc.RequestWithContext(ctx, action, b)

	if err != nil {
		return nil, err
	}

	response := make(map[string]any)
	err = json.Unmarshal(msg.Data, &response)

	if err != nil {
		return nil, err
	}

	return response, nil

}
