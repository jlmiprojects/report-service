// Package reportapi is the curated API report scripts call to reach Mongo and
// NATS. It is exported into the yaegi interpreter (see gen.go) so interpreted
// report scripts (scripts/*.go) get real Go control flow without ever seeing
// the raw mongo-driver/nats.go packages.
package reportapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"blueassetgroup.com/reports-service/repository"
	"blueassetgroup.com/reports-service/service"
	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// maxReportRows caps how many documents a single Find/Aggregate call may
// return, so a too-broad filter/pipeline can't exhaust memory rendering a
// report.
const maxReportRows = 50000

// Result is what Find/Aggregate return: JSON-friendly rows (ObjectID -> hex
// string, dates -> RFC3339 string) plus a count.
type Result struct {
	Data  []map[string]any
	Count int
}

// Option is one resolved lookup-parameter choice (see Options entry point).
type Option struct {
	Name  string
	Value string
}

// deps is wired once at startup (see Init) so Context/MongoHandle can reach
// the shared connection registry / NATS service call without every script
// needing to construct or receive them directly.
var deps struct {
	conns *repository.MongoConnections
	nats  *service.ServiceCall
}

// Init wires the shared Mongo connection registry and NATS service caller
// used by every Context this process creates. Call it once at startup before
// any report script runs.
func Init(conns *repository.MongoConnections, nats *service.ServiceCall) {
	deps.conns = conns
	deps.nats = nats
}

// Context is passed into every script's entry point (Run / Options).
type Context struct {
	// Query holds single-value URL query params.
	Query map[string]string
	// QueryAll holds every value for params that may repeat (e.g. CHECKBOX).
	QueryAll map[string][]string
	// ReportName is the report's Name, for logging / self-reference.
	ReportName string
	// Now is request time, injected so scripts don't need to call time.Now().
	Now time.Time
}

// Mongo returns a handle scoped to a named mongo_connections entry (blank =>
// "default").
func (c Context) Mongo(connection string) MongoHandle {
	return MongoHandle{connection: connection}
}

// Nats performs a NATS request/reply call: request is JSON-marshaled, the
// reply is JSON-unmarshaled into the returned value.
func (c Context) Nats(subject string, ttl time.Duration, request map[string]any) (any, error) {
	if deps.nats == nil {
		return nil, errors.New("reportapi: Nats called before Init")
	}
	return deps.nats.Generic(subject, ttl, request)
}

// MongoHandle scopes Find/Aggregate calls to one named Mongo connection.
type MongoHandle struct {
	connection string
}

// Find runs a find against database/collection. opts may set "projection",
// "sort", "limit", "skip" (same keys the old mongo.find data action used).
// filter/opts values may use Extended-JSON {"$oid": "<hex>"} / {"$date":
// "<rfc3339>"} to match ObjectID / date fields.
func (h MongoHandle) Find(database, collection string, filter, opts map[string]any) (Result, error) {
	client, def, err := h.client()
	if err != nil {
		return Result{}, err
	}

	dbName := database
	if dbName == "" {
		dbName = def.Database
	}
	if dbName == "" {
		return Result{}, errors.New("reportapi: no database (pass one to Find or configure the connection)")
	}
	if collection == "" {
		return Result{}, errors.New("reportapi: Find requires a collection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout(def))
	defer cancel()

	coll := client.Database(dbName).Collection(collection)

	findOpts := options.Find()
	if opts != nil {
		if p := asDoc(opts["projection"]); p != nil {
			findOpts.SetProjection(p)
		}
		if s := asDoc(opts["sort"]); s != nil {
			findOpts.SetSort(s)
		}
		if l, ok := reqInt64(opts, "limit"); ok {
			findOpts.SetLimit(l)
		}
		if sk, ok := reqInt64(opts, "skip"); ok {
			findOpts.SetSkip(sk)
		}
	}

	f := asDoc(mongoifyValue(filter))
	if f == nil {
		f = bson.M{}
	}

	cursor, err := coll.Find(ctx, f, findOpts)
	if err != nil {
		return Result{}, err
	}
	defer cursor.Close(ctx)

	return drain(ctx, cursor)
}

// Aggregate runs an aggregation pipeline against database/collection.
func (h MongoHandle) Aggregate(database, collection string, pipeline []any) (Result, error) {
	client, def, err := h.client()
	if err != nil {
		return Result{}, err
	}

	dbName := database
	if dbName == "" {
		dbName = def.Database
	}
	if dbName == "" {
		return Result{}, errors.New("reportapi: no database (pass one to Aggregate or configure the connection)")
	}
	if collection == "" {
		return Result{}, errors.New("reportapi: Aggregate requires a collection")
	}
	if len(pipeline) == 0 {
		return Result{}, errors.New("reportapi: Aggregate requires a non-empty pipeline")
	}

	ctx, cancel := context.WithTimeout(context.Background(), readTimeout(def))
	defer cancel()

	coll := client.Database(dbName).Collection(collection)

	cursor, err := coll.Aggregate(ctx, mongoifyValue(pipeline))
	if err != nil {
		return Result{}, err
	}
	defer cursor.Close(ctx)

	return drain(ctx, cursor)
}

func (h MongoHandle) client() (*mongo.Client, utils.MongoConnection, error) {
	if deps.conns == nil {
		return nil, utils.MongoConnection{}, errors.New("reportapi: Mongo called before Init")
	}
	return deps.conns.Connection(h.connection)
}

func readTimeout(def utils.MongoConnection) time.Duration {
	if def.ReadTimeout > 0 {
		return def.ReadTimeout
	}
	return 120 * time.Second
}

func drain(ctx context.Context, cursor *mongo.Cursor) (Result, error) {
	rows := make([]map[string]any, 0)
	for cursor.Next(ctx) {
		if len(rows) >= maxReportRows {
			return Result{}, fmt.Errorf("reportapi: query returned more than %d rows — narrow the filter", maxReportRows)
		}
		var doc map[string]any
		if err := cursor.Decode(&doc); err != nil {
			return Result{}, err
		}
		rows = append(rows, normaliseMongoValue(doc).(map[string]any))
	}
	if err := cursor.Err(); err != nil {
		return Result{}, err
	}
	return Result{Data: rows, Count: len(rows)}, nil
}

// ---- request helpers -------------------------------------------------------

func reqInt64(req map[string]any, key string) (int64, bool) {
	switch n := req[key].(type) {
	case float64:
		return int64(n), true
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	default:
		return 0, false
	}
}

func asDoc(v any) bson.M {
	switch x := v.(type) {
	case bson.M: // == primitive.M (type alias)
		return x
	case map[string]any:
		return bson.M(x)
	default:
		return nil
	}
}

// ---- value conversion ----------------------------------------------------

// normaliseMongoValue recursively converts driver BSON types into plain
// JSON-friendly values so html/template and yaegi scripts handle them:
// ObjectID -> hex string, DateTime/Timestamp/time.Time -> RFC3339 string,
// primitive.A/M/D -> []any / map[string]any.
func normaliseMongoValue(v any) any {
	switch x := v.(type) {
	case primitive.ObjectID:
		return x.Hex()
	case primitive.DateTime:
		return x.Time().UTC().Format(time.RFC3339)
	case primitive.Timestamp:
		return time.Unix(int64(x.T), 0).UTC().Format(time.RFC3339)
	case time.Time:
		return x.UTC().Format(time.RFC3339)
	case primitive.Decimal128:
		return x.String()
	case primitive.Binary:
		return x.Data
	case primitive.A:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = normaliseMongoValue(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = normaliseMongoValue(e)
		}
		return out
	case primitive.M:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = normaliseMongoValue(e)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = normaliseMongoValue(e)
		}
		return out
	case primitive.D:
		out := make(map[string]any, len(x))
		for _, e := range x {
			out[e.Key] = normaliseMongoValue(e.Value)
		}
		return out
	default:
		return v
	}
}

// mongoifyValue recursively rewrites a filter/pipeline so scripts can match
// typed fields with Extended-JSON wrappers:
//
//	{"$oid": "<24-hex>"}   -> primitive.ObjectID
//	{"$date": "<rfc3339>"} -> primitive.DateTime
func mongoifyValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		if len(x) == 1 {
			if s, ok := x["$oid"].(string); ok {
				if id, err := primitive.ObjectIDFromHex(s); err == nil {
					return id
				}
			}
			if s, ok := x["$date"].(string); ok {
				if t, err := time.Parse(time.RFC3339, s); err == nil {
					return primitive.NewDateTimeFromTime(t)
				}
			}
		}
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = mongoifyValue(e)
		}
		return out
	case primitive.M:
		return mongoifyValue(map[string]any(x))
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = mongoifyValue(e)
		}
		return out
	case primitive.A:
		return mongoifyValue([]any(x))
	default:
		return v
	}
}
