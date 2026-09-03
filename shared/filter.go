package utils

import (
	"encoding/json"
	"net/url"
)

type Filter struct {
	Value     string `json:"value"`
	MatchMode string `json:"matchMode"`
}

// {"first":0,"rows":10,"sortField":null,"sortOrder":null,"filters":{"global":{"value":null,"matchMode":"contains"}}}
type FindFilter struct {
	ID        string `json:"id"`
	ProfileId string `json:"profile_id"`
	GroupId   string `json:"group_id"`
	Owner     string `json:"owner"`
	Rows      int64  `json:"rows"`
	First     int64  `json:"first"`
	Search    string `json:"search"`
	Params    string `json:"params"`
	SortField string `json:"sortField"`
	SortOrder int    `json:"sortOrder"`
	// This will be used to exclude devices from the profile where
	ExcludeSubProfile string            `json:"exclude_sub_profile_id"`
	OwnerOnly         any               `json:"owner_only"`
	Filters           map[string]Filter `json:"filters"`
	VAS               any               `json:"vas"`
	VasFilter         string            `json:"vas_filter"`
	Driver            any               `json:"driver"`
	DriverKey         string            `json:"driver_key"` // this is the extract driver info from the endpoints
	OwnerDetail       any               `json:"owner_detail"`
	ShowLastPoint     any               `json:"show_lastpoint"`
	Status            string            `json:"status"`
}

func (fu *FindFilter) ParseParams(params string) error {

	j, err := url.QueryUnescape(params)

	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(j), fu)

	if err != nil {
		return err
	}

	return nil
}

/*func (t *FindFilter) UnmarshalJSON(b []byte) (err error) {

	dummy := make(map[string]interface{})
	newValue, _ := url.QueryUnescape(string(b))

	err := json.Unmarshal(b, &dummy)
	if err != nil {
		return err
	}
	return t.FromJSON([]byte(newValue))
}*/

func (f FindFilter) GetSearch() string {

	if f.Filters != nil {
		if v, ok := f.Filters["global"]; ok {
			return v.Value
		}
	}

	return ""

}

func (f *FindFilter) FromJSON(data []byte) error {

	err := json.Unmarshal(data, f)
	if err != nil {
		return err
	}

	return f.parseParams()

}

func (f *FindFilter) parseParams() error {

	if len(f.Params) > 0 {

		j, err := url.QueryUnescape(f.Params)

		if err != nil {
			return err
		}

		if len(j) == 0 {
			j = "{}"
		}

		err = json.Unmarshal([]byte(j), f)

		if err != nil {
			return err
		}

		if f.Rows == 0 {
			f.Rows = 1000
		}

	}

	return nil
}

func (u FindFilter) ToBytes() []byte {
	b, _ := json.Marshal(u)
	return b
}
