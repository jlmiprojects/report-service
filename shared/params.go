package utils

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
)

// This parses the params that is in url encoded json format
func ParseParams(params string, fu *FindFilter) error {

	j, err := url.QueryUnescape(params)

	if err != nil {
		return err
	}

	if len(j) == 0 {
		j = "{}"
	}

	/*if !strings.HasPrefix("{}", j) && !strings.HasSuffix("}", j) {
		return errors.New("invalid_json")
	}*/

	err = json.Unmarshal([]byte(j), fu)

	if err != nil {
		return err
	}

	if fu.Rows == 0 {
		fu.Rows = 1000
	}

	return nil
}

func GetAllBSONFieldsNames(model interface{}) []string {

	retval := make([]string, 0)
	val := reflect.ValueOf(model)
	for i := 0; i < val.Type().NumField(); i++ {

		if len(val.Type().Field(i).Tag.Get("bson")) > 0 && val.Type().Field(i).Tag.Get("bson") != "-" {

			parts := strings.Split(val.Type().Field(i).Tag.Get("bson"), ",")

			retval = append(retval, parts[0])
		}
	}

	return retval

}
