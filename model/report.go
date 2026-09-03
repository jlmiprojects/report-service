package model

import (
	"encoding/json"
	"fmt"
	"strconv"

	utils "blueassetgroup.com/reports-service/shared"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FontStyle struct {
	Bold      bool    `json:"bold"`
	Italic    bool    `json:"italic"`
	Underline string  `json:"underline"`
	Family    string  `json:"family"`
	Size      float64 `json:"size"`
	Strike    bool    `json:"strike"`
	Color     string  `json:"color"`
}

type BorderStyle struct {
	Top    bool   `json:"top"`
	Left   bool   `json:"left"`
	Bottom bool   `json:"bottom"`
	Right  bool   `json:"right"`
	Color  string `json:"color"`
	Style  int    `json:"style"`
}

type FillStyle struct {
	Type         string   `json:"type"`
	Pattern      int      `json:"pattern"`
	Color        []string `json:"color"`
	Shading      int      `json:"shading"`
	Transparency int      `json:"transparency"`
}

type ExcelReporCell struct {
	Value  any            `json:"value"`
	Type   string         `json:"type"` // string, int, float, date
	Border []*BorderStyle `json:"border,omitempty"`
	Font   *FontStyle     `json:"font"`
	Fill   *FillStyle     `json:"fill"`
}

type PageLayout struct {
	Orientation string `json:"orientation"`
	FitToWidth  bool   `json:"fit_to_width"`
}

type ExcelReport struct {
	Header     string             `json:"header"`
	Rows       [][]ExcelReporCell `json:"rows"`
	Footer     string             `json:"footer"`
	PageLayout PageLayout         `json:"page_layout"`
}

func (er *ExcelReport) FromJSON(b []byte) error {

	return json.Unmarshal(b, er)

}

type Report struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Handler      string             `json:"handler" bson:"handler"`
	Description  string             `json:"description"`
	Parameters   []ReportParams     `json:"parameters"`
	ACI          string             `json:"aci"`
	TemplateName string             `json:"template"`
	Management   bool               `json:"management" bson:"management"`
	DataActions  []*DataActions     `json:"data_actions" bson:"data_actions"`
}

type ReportParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Metadata    any    `json:"metadata"`
}

const (
	DATA_ACTION_NATS  = "nats"
	DATA_ACTION_JS    = "js"
	DATA_ACTION_MONGO = "mongo"
)

type DataActions struct {
	Name string `json:"name" bson:"name"`
	Type string `json:"type"`
	// Action is the NATS subject (nats), script name (js) or mongo operation
	// (mongo.find / mongo.aggregate) to run.
	Action string `json:"action" bson:"action"`
	// Connection selects a named entry from config `mongo_connections` for a
	// `mongo` action. Blank => "default".
	Connection string         `json:"connection" bson:"connection"`
	Request    map[string]any `json:"request" bson:"request"`
	TTL        string         `json:"ttl" bson:"ttl"`
}

type FindAllReportsRequest struct {
	Management bool
}

func (m *FindAllReportsRequest) UnmarshalJSON(data []byte) error {
	var temp map[string]any
	if err := json.Unmarshal(data, &temp); err == nil {

		if temp["management"] == nil {
			m.Management = false
			return nil
		}

		if s, ok := temp["management"].(string); ok {
			parsedBool, err := strconv.ParseBool(s)
			if err != nil {
				return fmt.Errorf("cannot parse string '%s' into bool: %w", s, err)
			}
			m.Management = parsedBool
		} else {
			m.Management = temp["management"].(bool)
		}

	}

	return nil
}

type FindAllReportsResult struct {
	utils.Result
	Reports []*Report `json:"reports"`
}

type ReportSchedule struct {
	ID         primitive.ObjectID `json:"id" bson:"_id"`
	ProfileId  primitive.ObjectID `json:"profile_id" validate:"nonzero"`
	ReportId   primitive.ObjectID `json:"report_id" validate:"nonzero"`
	Parameters map[string]any     `json:"parameters"`
	StartTime  string             `json:"start_time" bson:"start_time" validate:"nonzero"`
	//StartDate      string             `json:"start_date" bson:"-" validate:"nonzero"`
	//MongoStartDate time.Time          `bson:"startdate" json:"-"`
	Monday        bool `json:"monday"`
	Tuesday       bool `json:"tuesday"`
	Wednesday     bool `json:"wednesday"`
	Thursday      bool `json:"thursday"`
	Friday        bool `json:"friday"`
	Saturday      bool `json:"saturday"`
	Sunday        bool `json:"sunday"`
	RepeatDaily   bool `json:"repeat_daily"`
	RepeatMonthly bool `json:"repeat_monthly"`
	RepeatWeekly  bool `json:"repeat_weekly"`
}

func (u *ReportSchedule) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u ReportSchedule) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}

type AddReportScheduleResponse struct {
	utils.Result
	ID string `json:"id"`
}

type FindReportScheduleRequest struct {
	ProfileId string `json:"profile_id"`
	ReportId  string `json:"report_id"`
}

func (u *FindReportScheduleRequest) FromJSON(data []byte) error {
	return json.Unmarshal(data, u)
}

type FindReportScheduleResponse struct {
	utils.Result
	ProfileId string          `json:"profile_id"`
	ReportId  string          `json:"report_id"`
	Schedule  *ReportSchedule `json:"schedule"`
}

func (u FindReportScheduleRequest) Validate() []*utils.ValidationError {
	return utils.GetValidate().Validate(u)
}
