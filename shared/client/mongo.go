package client

import (
	"context"
	"errors"
	"time"

	"blueassetgroup.com/reports-service/shared/models/mongo"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
)

func MongoCount(nc *nats.Conn, database string, collection string, filter map[string]any) (int64, error) {

	request := mongo.MongoCountRequest{}

	request.Collection = collection
	request.Database = database

	if filter != nil {
		request.Filter = filter
	}

	j := request.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, "mongo.count", j)

	if err != nil {
		return 0, err
	}

	response := new(mongo.MongoCountResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return 0, err
	}

	if response.StatusCode != 200 {
		return 0, errors.New(response.Error)
	}

	return response.Count, nil

}

func MongoFind(nc *nats.Conn, database string, collection string, filter bson.M, projection bson.M, limit int64) ([]map[string]any, int64, error) {

	request := mongo.MongoFindRequest{}

	request.Collection = collection
	request.Database = database

	if filter != nil {
		request.Filter = filter
	}

	if projection != nil {
		request.Projection = projection
	}

	if limit == 0 {
		limit = 1000
	}
	request.Limit = limit

	j := request.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, "mongo.find", j)

	if err != nil {
		return nil, 0, err
	}

	response := new(mongo.MongoFindResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode != 200 {
		return nil, 0, errors.New(response.Error)
	}

	return response.Data, int64(response.Count), nil

}

func MongoAggregate(nc *nats.Conn, database string, collection string, pipelne []bson.M) ([]map[string]any, int64, error) {

	request := mongo.MongoAggregateRequest{}

	request.Collection = collection
	request.Database = database
	request.Pipeline = pipelne

	j := request.ToBytes()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)

	defer cancel()

	msg, err := nc.RequestWithContext(ctx, "mongo.aggregate", j)

	if err != nil {
		return nil, 0, err
	}

	response := new(mongo.MongoFindResponse)

	err = response.FromJSON(msg.Data)

	if err != nil {
		return nil, 0, err
	}

	if response.StatusCode != 200 {
		return nil, 0, errors.New(response.Error)
	}

	return response.Data, int64(response.Count), nil

}
