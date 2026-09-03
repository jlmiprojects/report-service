package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"blueassetgroup.com/reports-service/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// maxReportRows caps how many documents a single mongo DataAction may return,
// so a too-broad filter/pipeline can't exhaust memory rendering a report.
const maxReportRows = 50000

// RunMongoDataAction executes a `mongo` report DataAction against the named
// connection (action.Connection, blank => "default") and returns the dataset in
// the same shape a NATS list reply uses: {"data": [...docs...], "count": N}, so
// report templates and JS scripts consume it as data.<name>.data.
//
// action.Request keys:
//
//	mongo.find      – collection, [database], [filter], [projection], [sort], [limit], [skip]
//	mongo.aggregate – collection, [database], pipeline
//
// find filters may use Extended-JSON {"$oid": "<hex>"} / {"$date": "<rfc3339>"}
// to match ObjectID / date fields; aggregate pipelines use $toObjectId etc. and
// are passed through untouched.
func (m *MongoConnections) RunMongoDataAction(action *model.DataActions) (map[string]any, error) {
	client, def, err := m.connection(action.Connection)
	if err != nil {
		return nil, err
	}

	req := action.Request
	if req == nil {
		req = map[string]any{}
	}

	collName := reqString(req, "collection")
	if collName == "" {
		return nil, errors.New("mongo data action requires a 'collection'")
	}

	dbName := reqString(req, "database")
	if dbName == "" {
		dbName = def.Database
	}
	if dbName == "" {
		return nil, errors.New("mongo data action has no database (set 'database' on the request or the connection)")
	}

	timeout := def.ReadTimeout
	if action.TTL != "" {
		if d, perr := time.ParseDuration(action.TTL); perr == nil {
			timeout = d
		}
	}
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	coll := client.Database(dbName).Collection(collName)

	op := action.Action
	if op == "" || op == "mongo" {
		op = "mongo.find"
	}

	var cursor *mongo.Cursor
	switch op {
	case "mongo.find":
		findOpts := options.Find()
		if p := asDoc(req["projection"]); p != nil {
			findOpts.SetProjection(p)
		}
		if s := asDoc(req["sort"]); s != nil {
			findOpts.SetSort(s)
		}
		if l, ok := reqInt64(req, "limit"); ok {
			findOpts.SetLimit(l)
		}
		if sk, ok := reqInt64(req, "skip"); ok {
			findOpts.SetSkip(sk)
		}
		filter := asDoc(mongoifyValue(req["filter"]))
		if filter == nil {
			filter = bson.M{}
		}
		cursor, err = coll.Find(ctx, filter, findOpts)
	case "mongo.aggregate":
		pipeline := asSlice(req["pipeline"])
		if pipeline == nil {
			return nil, errors.New("mongo.aggregate requires a 'pipeline' array")
		}
		cursor, err = coll.Aggregate(ctx, pipeline)
	default:
		return nil, fmt.Errorf("unknown mongo action %q (use mongo.find or mongo.aggregate)", action.Action)
	}
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	rows := make([]any, 0)
	for cursor.Next(ctx) {
		if len(rows) >= maxReportRows {
			return nil, fmt.Errorf("mongo data action returned more than %d rows — narrow the filter", maxReportRows)
		}
		var doc map[string]any
		if derr := cursor.Decode(&doc); derr != nil {
			return nil, derr
		}
		rows = append(rows, normaliseMongoValue(doc))
	}
	if cerr := cursor.Err(); cerr != nil {
		return nil, cerr
	}

	return map[string]any{"data": rows, "count": len(rows)}, nil
}

// ---- request helpers -------------------------------------------------------

func reqString(req map[string]any, key string) string {
	if v, ok := req[key].(string); ok {
		return v
	}
	return ""
}

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

func asSlice(v any) []any {
	switch x := v.(type) {
	case []any:
		return x
	case primitive.A: // == bson.A (type alias)
		return []any(x)
	default:
		return nil
	}
}

// ---- value conversion ----------------------------------------------------

// normaliseMongoValue recursively converts driver BSON types into plain
// JSON-friendly values so html/template and the sobek JS VM handle them:
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

// mongoifyValue recursively rewrites a find filter so report authors can match
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
