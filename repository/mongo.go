package repository

import (
	"context"

	"blueassetgroup.com/reports-service/model"
	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	client *mongo.Client
	Config *utils.Config
}

func NewMongo(config *utils.Config) (*MongoRepository, error) {

	instance := new(MongoRepository)

	instance.Config = config

	var err error

	bsonOpts := &options.BSONOptions{
		UseJSONStructTags: true,
		NilSliceAsEmpty:   true,
		// DefaultDocumentM makes any interface{}-typed field (e.g.
		// ReportParams.Metadata, ReportSchedule.Parameters) decode
		// subdocuments as primitive.M (map[string]any) instead of the
		// driver's default primitive.D (an ordered []{Key,Value} list).
		// Without it, a JSON-marshaled Metadata like
		// {"options":[{"name":...,"value":...}]} comes out as
		// [{"Key":"options","Value":[...]}] over NATS — broker-portal's
		// Parameter.Options() then can't find "options" and silently
		// renders zero options for SELECT/RADIO/CHECKBOX, and
		// model.ParseLookupMetadata fails to unmarshal LOOKUP metadata.
		DefaultDocumentM: true,
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(config.Mongo.Uri).SetServerAPIOptions(serverAPI).SetBSONOptions(bsonOpts)

	ctx, cancel := context.WithTimeout(context.Background(), config.Mongo.ConnectTimeout)
	defer cancel()

	instance.client, err = mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	err = instance.client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	return instance, nil

}

func (r MongoRepository) FindAll(management bool) ([]*model.Report, error) {

	ctx, cancel := context.WithTimeout(context.Background(), r.Config.Mongo.ReadTimeout)
	defer cancel()

	db := r.client.Database(r.Config.Mongo.Database)
	coll := db.Collection(r.Config.Mongo.ReportTable)

	c, err := coll.Find(ctx, bson.M{"management": management}, nil)
	if err != nil {
		return nil, err
	}

	reports := make([]*model.Report, 0)
	for c.Next(context.TODO()) {
		var result = new(model.Report)
		if err := c.Decode(result); err != nil {
			return nil, err
		}

		//result.ID = result.MongoID.Hex()
		reports = append(reports, result)
	}

	return reports, nil
}

func (r MongoRepository) FindByName(name string) (*model.Report, error) {

	ctx, cancel := context.WithTimeout(context.Background(), r.Config.Mongo.ReadTimeout)
	defer cancel()

	db := r.client.Database(r.Config.Mongo.Database)
	coll := db.Collection(r.Config.Mongo.ReportTable)

	c := coll.FindOne(ctx, bson.M{"name": name}, nil)
	if c.Err() != nil {
		return nil, c.Err()
	}

	var result = new(model.Report)
	if err := c.Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r MongoRepository) FindByID(id string) (*model.Report, error) {

	ctx, cancel := context.WithTimeout(context.Background(), r.Config.Mongo.ReadTimeout)
	defer cancel()

	db := r.client.Database(r.Config.Mongo.Database)
	coll := db.Collection(r.Config.Mongo.ReportTable)

	_reportId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	c := coll.FindOne(ctx, bson.M{"_id": _reportId}, nil)
	if c.Err() != nil {
		return nil, c.Err()
	}

	var result = new(model.Report)
	if err := c.Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r MongoRepository) AddReportSchedule(schedule *model.ReportSchedule) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), r.Config.Mongo.ReadTimeout)
	defer cancel()

	db := r.client.Database(r.Config.Mongo.Database)
	coll := db.Collection(r.Config.Mongo.ReportScheduleTable)

	schedule.ID = primitive.NewObjectID()

	/*t, err := time.ParseInLocation("2006-01-02 15:04", schedule.StartDate, time.Local)
	if err != nil {
		return "", errors.WithMessage(err, "Failed tp parse start ")
	}
	schedule.MongoStartDate = t*/

	_, err := coll.InsertOne(ctx, schedule)
	if err != nil {
		return "", err
	}

	return schedule.ID.Hex(), nil
}

func (r MongoRepository) FindReportSchedule(profileId string, reportId string) (*model.ReportSchedule, error) {

	ctx, cancel := context.WithTimeout(context.Background(), r.Config.Mongo.ReadTimeout)
	defer cancel()

	db := r.client.Database(r.Config.Mongo.Database)
	coll := db.Collection(r.Config.Mongo.ReportScheduleTable)

	_profileId, err := primitive.ObjectIDFromHex(profileId)
	if err != nil {
		return nil, err
	}

	_reportId, err := primitive.ObjectIDFromHex(reportId)
	if err != nil {
		return nil, err
	}

	c := coll.FindOne(ctx, bson.M{"report_id": _reportId, "profile_id": _profileId}, nil)
	if c.Err() != nil {
		return nil, err
	}

	data := new(model.ReportSchedule)

	err = c.Decode(data)

	if err != nil {
		return nil, err
	}

	//data.StartDate = data.MongoStartDate.Format("2006-01-02 15:04")

	if data.Parameters == nil {
		data.Parameters = make(map[string]any)
	}

	return data, nil
}

func (r MongoRepository) FindAllReportSchedule() ([]*model.ReportSchedule, error) {

	ctx, cancel := context.WithTimeout(context.Background(), r.Config.Mongo.ReadTimeout)
	defer cancel()

	db := r.client.Database(r.Config.Mongo.Database)
	coll := db.Collection(r.Config.Mongo.ReportScheduleTable)

	c, err := coll.Find(ctx, bson.M{}, nil)
	if err != nil {
		return nil, err
	}

	data := make([]*model.ReportSchedule, 0)

	err = c.All(ctx, &data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r MongoRepository) FinddReportScheduleByTime(targetTime string) ([]*model.ReportSchedule, error) {

	ctx, cancel := context.WithTimeout(context.Background(), r.Config.Mongo.ReadTimeout)
	defer cancel()

	db := r.client.Database(r.Config.Mongo.Database)
	coll := db.Collection(r.Config.Mongo.ReportScheduleTable)

	filter := bson.M{"start_time": targetTime}

	c, err := coll.Find(ctx, filter, nil)
	if err != nil {
		return nil, err
	}

	data := make([]*model.ReportSchedule, 0)

	err = c.All(ctx, &data)

	if err != nil {
		return nil, err
	}

	return data, nil
}
