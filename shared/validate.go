package utils

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"gopkg.in/validator.v2"
)

var validate *Validator
var validateOnce sync.Once

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Validator struct {
}

func GetValidate() *Validator {

	validateOnce.Do(func() {
		validate = &Validator{}

		validator.SetValidationFunc("date", date)
	})

	return validate

}

func (v Validator) Validate(p any) []*ValidationError {

	err := validator.Validate(p)

	if err == nil {
		return nil
	}

	errs := err.(validator.ErrorMap)

	retval := make([]*ValidationError, 0)
	for k, v := range errs {
		oneError := new(ValidationError)
		oneError.Field = getJSONFieldName(p, k)
		oneError.Message = v.Error()

		if oneError.Message == "zero value" {
			oneError.Message = "required"
		}

		retval = append(retval, oneError)
	}

	return retval
}

func date(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("data only validates strings")
	}

	// Parse to yyyy-MM-dd HH:mm"
	_, err := time.Parse(st.String(), "2006-01-02 15:04")

	if err != nil {
		return err

	}
	return nil
}

func getJSONFieldName(d any, fieldName string) string {

	val := reflect.ValueOf(d)

	f, found := val.Type().FieldByName(fieldName)

	if !found {
		return fieldName
	}

	tag := f.Tag.Get("json")

	if tag == "" {
		return f.Name
	}

	if tag == "-" {
		return fieldName
	}

	if i := strings.Index(tag, ","); i != -1 {
		if i == 0 {
			return f.Name
		} else {
			return tag[:i]
		}
	}

	return tag
}

func MakeError(e, v string) *ValidationError {

	res := ValidationError{}

	res.Field = e
	res.Message = v

	return &res

}
