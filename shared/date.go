package utils

import (
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type CustomTime struct {
	time.Time
}

func (t *CustomTime) Convert(_t string) error {
	date, err := time.ParseInLocation("2006-01-02 15:04", _t, time.Local)

	if err != nil {
		return err
	}
	t.Time = date

	return nil
}

func (t *CustomTime) UnmarshalJSON(b []byte) (err error) {

	_t := string(b)

	_t = strings.ReplaceAll(_t, "\"", "")
	date, err := time.ParseInLocation("2006-01-02 15:04", _t, time.Local)
	if err != nil {
		return err
	}
	t.Time = date
	return
}

func (t CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.Time.Format("2006-01-02 15:04") + `"`), nil
}

type RealtimeTime struct {
	time.Time
}

func (t *RealtimeTime) UnmarshalJSON(b []byte) (err error) {

	_t := string(b)

	_t = strings.ReplaceAll(_t, "\"", "")

	date, err := time.ParseInLocation("2006-01-02 15:04:05", _t, time.Local)
	if err != nil {
		return err
	}
	t.Time = date
	return
}

func (t *RealtimeTime) MarshalJSON() ([]byte, error) {

	formatted := t.Format("2006-01-02 15:04:05")

	jsonStr := "\"" + formatted + "\""
	return []byte(jsonStr), nil

}

func (v *RealtimeTime) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	rv := bson.RawValue{
		Type:  t,
		Value: b,
	}

	var res time.Time
	if err := rv.Unmarshal(&res); err != nil {
		return err
	}

	v.Time = res

	return nil
}

func (v RealtimeTime) String() string {

	return v.Time.Format("2006-01-02 15:04:05")

}

//type CustomDate time.Time

type CustomDate struct {
	time.Time
}

func (t *CustomDate) UnmarshalJSON(b []byte) (err error) {

	_t := string(b)

	_t = strings.ReplaceAll(_t, "\"", "")

	if len(_t) == 0 {
		t.Time = time.Now()
		return nil
	}

	date, err := time.ParseInLocation("2006-01-02", _t, time.Local)
	if err != nil {
		return err
	}
	t.Time = date
	return
}

func (t CustomDate) MarshalJSON() ([]byte, error) {

	formatted := t.Format("2006-01-02")

	jsonStr := "\"" + formatted + "\""
	return []byte(jsonStr), nil

}

func (v CustomDate) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(v)
}

func (v *CustomDate) UnmarshalBSONValue(t bsontype.Type, b []byte) error {
	rv := bson.RawValue{
		Type:  t,
		Value: b,
	}

	var res time.Time
	if err := rv.Unmarshal(&res); err != nil {
		return err
	}

	v.Time = res

	return nil
}

func (v CustomDate) String() string {

	return v.Format("2006-01-02")

}

func NewCustomeDate(_t string) (*CustomDate, error) {

	date, err := time.ParseInLocation("2006-01-02", _t, time.Local)
	if err != nil {
		return nil, err
	}
	return &CustomDate{date}, nil
}

// yyyy-MM-dd HH:mm:ss
func TimestampToLocal(format string, _t string) string {

	location, err := time.LoadLocation("Africa/Johannesburg")

	if err != nil {
		return err.Error()
	}

	date, err := time.Parse(format, _t)

	if err != nil {
		return err.Error()
	}

	_date := date.In(location)

	return _date.Format(format)
}
