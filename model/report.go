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
	// Excel is set when the report also has an <name>.csv template and can be
	// rendered as an .xlsx attachment (GET /run?...&type=csv). Consumers use it
	// to decide whether to offer an Excel download.
	Excel bool `json:"excel" bson:"excel"`
	// Script names a .go file under config ScriptDir (e.g. "clients" ->
	// <script_dir>/clients.go), interpreted via yaegi at request time to
	// build the report's data map (see handlers.ScriptRunner).
	Script string `json:"script" bson:"script"`
}

/*
type can be one of the following:

 1. DATE -> Display a datetime selector
 2. SELECT -> static dropdown. METADATA: {"options":[{"name":"One","value":"1"}]}
 3. RADIO -> static radio-button group. Same METADATA shape as SELECT.
 4. CHECKBOX -> static multi-select checkbox group (submits repeated query
    params under Name). Same METADATA shape as SELECT. Distinct from BOOL,
    which is a single yes/no checkbox.
 5. LOOKUP -> options resolved server-side by a yaegi script instead of being
    static. METADATA is a LookupMetadata (script + display) — see
    LookupMetadata / ParseLookupMetadata.
 6. TEXT -> Normal text entry. Regexp, if set, is an HTML5 pattern the value
    must match. "string" is an equivalent legacy value — both render the
    same way (see broker-portal's reportParamField default case).
 7. NUMBER -> A number
 8. BOOL -> A single yes/no checkbox
*/
type ReportParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Metadata    any    `json:"metadata"`
	// Regexp, when set, is an HTML5 pattern a STRING parameter's value must
	// match (checked both client-side via the rendered form and server-side
	// on /run). Ignored for every other Type.
	Regexp string `json:"regexp"`
}

const (
	PARAM_TYPE_LOOKUP   = "lookup"
	PARAM_TYPE_SELECT   = "select"
	PARAM_TYPE_RADIO    = "radio"
	PARAM_TYPE_CHECKBOX = "checkbox"
	PARAM_TYPE_TEXT     = "text"
	PARAM_TYPE_NUMBER   = "number"
	PARAM_TYPE_DATE     = "date"
	PARAM_TYPE_BOOL     = "bool"
)

// Option is one choice in a SELECT/RADIO/CHECKBOX parameter's static options,
// or one row of a resolved LOOKUP's dynamic options.
type Option struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// LookupMetadata is the Metadata shape for a "lookup" parameter: it names a
// yaegi script (same reportapi.Context input a report Script gets) whose
// Options(ctx) entry point returns the dropdown's options directly.
type LookupMetadata struct {
	Script string `json:"script"`
	// Display selects how the resolved options render: "select" (default),
	// "radio" or "checkbox".
	Display string `json:"display,omitempty"`
}

// ParseLookupMetadata round-trips a "lookup" parameter's Metadata (decoded
// from JSON/BSON as `any`) into a LookupMetadata.
func ParseLookupMetadata(metadata any) (*LookupMetadata, error) {
	b, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	var lm LookupMetadata
	if err := json.Unmarshal(b, &lm); err != nil {
		return nil, err
	}
	return &lm, nil
}

// ParamOptionsRequest asks for one lookup parameter's resolved options,
// optionally scoped by other already-known parameter values (query).
type ParamOptionsRequest struct {
	Report    string         `json:"report"`
	Parameter string         `json:"parameter"`
	Query     map[string]any `json:"query"`
}

// ParamOptionsResult is the reply to a ParamOptionsRequest.
type ParamOptionsResult struct {
	utils.Result
	Options []Option `json:"options"`
	Display string   `json:"display"`
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
