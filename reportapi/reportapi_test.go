package reportapi

import (
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNormaliseMongoValue(t *testing.T) {
	oid := primitive.NewObjectID()
	when := time.Date(2025, 3, 1, 8, 30, 0, 0, time.UTC)

	in := map[string]any{
		"_id":     oid,
		"created": primitive.NewDateTimeFromTime(when),
		"nested": primitive.M{
			"owner": oid,
			"tags":  primitive.A{"a", oid, primitive.M{"deep": oid}},
		},
		"plain": "keep",
		"n":     int64(42),
	}

	got := normaliseMongoValue(in).(map[string]any)

	if got["_id"] != oid.Hex() {
		t.Errorf("_id = %v, want %s", got["_id"], oid.Hex())
	}
	if got["created"] != when.Format(time.RFC3339) {
		t.Errorf("created = %v, want %s", got["created"], when.Format(time.RFC3339))
	}
	if got["plain"] != "keep" || got["n"] != int64(42) {
		t.Errorf("plain values mutated: %#v", got)
	}

	nested := got["nested"].(map[string]any)
	if nested["owner"] != oid.Hex() {
		t.Errorf("nested.owner not normalised: %v", nested["owner"])
	}
	tags := nested["tags"].([]any)
	if tags[1] != oid.Hex() {
		t.Errorf("nested.tags[1] not normalised: %v", tags[1])
	}
	if tags[2].(map[string]any)["deep"] != oid.Hex() {
		t.Errorf("nested.tags[2].deep not normalised: %v", tags[2])
	}
}

func TestMongoifyValue_OID(t *testing.T) {
	oid := primitive.NewObjectID()

	filter := map[string]any{
		"_id":   map[string]any{"$oid": oid.Hex()},
		"owner": map[string]any{"$oid": oid.Hex()},
		"name":  "unchanged",
		"$or": []any{
			map[string]any{"status": "active"},
			map[string]any{"parent": map[string]any{"$oid": oid.Hex()}},
		},
	}

	got := mongoifyValue(filter).(map[string]any)

	if id, ok := got["_id"].(primitive.ObjectID); !ok || id != oid {
		t.Errorf("_id not unwrapped to ObjectID: %#v", got["_id"])
	}
	if got["name"] != "unchanged" {
		t.Errorf("name mutated: %v", got["name"])
	}
	or := got["$or"].([]any)
	if _, ok := or[1].(map[string]any)["parent"].(primitive.ObjectID); !ok {
		t.Errorf("$or[1].parent not unwrapped: %#v", or[1])
	}
}

func TestMongoifyValue_Date(t *testing.T) {
	when := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)

	got := mongoifyValue(map[string]any{
		"created": map[string]any{"$date": when.Format(time.RFC3339)},
	}).(map[string]any)

	dt, ok := got["created"].(primitive.DateTime)
	if !ok {
		t.Fatalf("created not unwrapped to DateTime: %#v", got["created"])
	}
	if !dt.Time().UTC().Equal(when) {
		t.Errorf("created = %v, want %v", dt.Time().UTC(), when)
	}
}

func TestMongoifyValue_BadOIDLeftAlone(t *testing.T) {
	in := map[string]any{"_id": map[string]any{"$oid": "not-a-hex"}}
	got := mongoifyValue(in).(map[string]any)
	// stays a nested map — driver will surface the error, not a silent match
	if !reflect.DeepEqual(got["_id"], map[string]any{"$oid": "not-a-hex"}) {
		t.Errorf("bad $oid should be left as-is, got %#v", got["_id"])
	}
}

func TestReqInt64(t *testing.T) {
	cases := []struct {
		in   any
		want int64
		ok   bool
	}{
		{float64(10), 10, true},
		{int(7), 7, true},
		{int32(3), 3, true},
		{int64(99), 99, true},
		{"5", 0, false},
		{nil, 0, false},
	}
	for _, c := range cases {
		got, ok := reqInt64(map[string]any{"limit": c.in}, "limit")
		if got != c.want || ok != c.ok {
			t.Errorf("reqInt64(%T %v) = (%d,%v), want (%d,%v)", c.in, c.in, got, ok, c.want, c.ok)
		}
	}
}
