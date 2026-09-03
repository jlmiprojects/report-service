package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"blueassetgroup.com/reports-service/model"
	"blueassetgroup.com/reports-service/repository"
	"blueassetgroup.com/reports-service/service"
	utils "blueassetgroup.com/reports-service/shared"
	"github.com/grafana/sobek"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"

	_ "image/png"
)

type ReportHandler struct {
	csvRender   *template.Template
	serviceCall *service.ServiceCall
	conns       *repository.MongoConnections
	funcMap     template.FuncMap
	scriptDir   string
	templateDir string
}

func NewReportHandler(serviceCall *service.ServiceCall, csvRender *template.Template, config *utils.Config, funcMap template.FuncMap, conns *repository.MongoConnections) *ReportHandler {

	/*funcMap := template.FuncMap{
		"parseCustomTime": func(timeStr string) (*utils.CustomTime, error) {

			customTime := utils.CustomTime{}

			err := customTime.Convert(timeStr)

			if err != nil {
				return nil, err
			}
			return &customTime, nil
		},
		"split": func(list string, sep string) []string {
			return strings.Split(list, sep)
		},
		"convertToLocal": func(_t string) string {

			date, err := time.ParseInLocation("2006-01-02T15:04:05", _t, time.Local)

			if err != nil {
				return err.Error()
			}

			return date.Format("2006-01-02 15:04:05")
		},
		"formatWorkingTime": func(a any) string {

			totalMinutes := cast.ToInt(a)

			hours := totalMinutes / 60
			minutes := totalMinutes % 60
			return fmt.Sprintf("%d:%02d", hours, minutes)

		},
		"mod": func(a, b int) int {
			return a % b
		},
	}*/

	return &ReportHandler{csvRender: csvRender, serviceCall: serviceCall, conns: conns, funcMap: funcMap, scriptDir: *config.ScriptDir, templateDir: *config.TemplateDir}
}

func (self ReportHandler) buildCSV(c echo.Context, data any, downLoadFileName string) error {

	fileName := fmt.Sprintf("/tmp/%s_%s.csv", c.QueryParam("profile_id"), c.QueryParam("name"))

	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	if err != nil {
		slog.Error("Error", "error", err)
		return c.Render(200, "error.html", fmt.Sprintf("Failed to open file %s", err.Error()))
	}

	var buf bytes.Buffer

	err = self.csvRender.ExecuteTemplate(&buf, c.QueryParam("name")+".csv", data)

	if err != nil {
		slog.Error("Error", "error", err)
		return c.Render(200, "error.html", fmt.Sprintf("Failed to render %s", err.Error()))
	}

	//csvreader := csv.NewReader(&buf)
	//csvreader.LazyQuotes = true

	slog.Info("CSV", "csv", buf.String())

	ef := excelize.NewFile()

	index, err := ef.NewSheet(c.QueryParam("name"))
	if err != nil {
		slog.Error("Error", "error", err)
		return c.Render(200, "error.html", fmt.Sprintf("Failed to create excell sheet %s", err.Error()))
	}

	ef.DeleteSheet("Sheet1")

	v := false
	margin := 1.5
	leftmargin := 0.5

	/*orientation := "portrait"
	fitToWidth := 1

	ef.SetPageLayout(c.QueryParam("name"), &excelize.PageLayoutOptions{
		Orientation: &orientation,
		FitToWidth:  &fitToWidth,
	})*/

	ef.SetSheetProps(c.QueryParam("name"), &excelize.SheetPropsOptions{FitToPage: &v})
	ef.SetPageMargins(c.QueryParam("name"), &excelize.PageLayoutMarginsOptions{
		Top:    &margin,
		Bottom: &margin,
		Left:   &leftmargin,
	})

	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "0", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	excell := new(model.ExcelReport)

	d := strings.ReplaceAll(buf.String(), "\n", "")

	err = excell.FromJSON([]byte(d))
	if err != nil {
		return err
	}

	fileBytes, err := os.ReadFile(self.templateDir + "/logo.png")

	if err == nil {
		err = ef.AddHeaderFooterImage(c.QueryParam("name"), &excelize.HeaderFooterImageOptions{
			Position:  excelize.HeaderFooterImagePositionRight,
			Extension: ".png",
			Width:     "120pt",
			Height:    "30pt",
			File:      fileBytes,
			IsFooter:  false,

			//FirstPage: true,
		})

		if err != nil {
			slog.Error("Failed to add images header ", "error", err)
		}
	}

	err = ef.SetHeaderFooter(c.QueryParam("name"), &excelize.HeaderFooterOptions{
		DifferentFirst:   false,
		DifferentOddEven: false,
		OddHeader:        excell.Header,
		OddFooter:        excell.Footer,
		AlignWithMargins: &v,
		ScaleWithDoc:     &v,
	})

	if err != nil {
		slog.Error("Failed to add footer and headers ", "error", err)
	}
	colWidths := make(map[int]float64)

	for rowIndex, row := range excell.Rows {
		for columnIndex, cell := range row {
			colRow := fmt.Sprintf("%s%d", columns[columnIndex], rowIndex+1)
			//colName, _ := excelize.ColumnNumberToName(columnIndex + 1)

			addCell(ef, colRow, c.QueryParam("name"), &cell)

			var val string
			if cell.Value != nil {
				val = fmt.Sprintf("%v", cell.Value)
			}
			width := float64(len(val)) + 2
			if width > colWidths[columnIndex] {
				colWidths[columnIndex] = width
			}
		}
	}

	for columnIndex, width := range colWidths {
		colName, _ := excelize.ColumnNumberToName(columnIndex + 1)
		ef.SetColWidth(c.QueryParam("name"), colName, colName, width)
	}

	ef.SetActiveSheet(index)

	_, err = ef.WriteTo(f)

	if err != nil {
		slog.Error("Error", "error", err)
		return c.Render(200, "error.html", fmt.Sprintf("Failed to create excell sheet %s", err.Error()))
	}

	return c.Attachment(fileName, downLoadFileName+".xlsx")

}

func (rh ReportHandler) runScript(name string, data map[string]any, query map[string]any) error {

	logger := slog.With("function", "runScript")

	vm := sobek.New()

	slog.Info("runScript", "data", data, "query", query)

	vm.Set("data", data)
	vm.Set("query", query)

	vm.Set("console", map[string]interface{}{
		"log": func(call sobek.FunctionCall) sobek.Value {
			logger.Info("JS Console Log", "args", call.Arguments)
			return sobek.Undefined()
		},
	})

	logger.Info("Params", "name", name, "dir", rh.scriptDir)

	b, err := os.ReadFile(fmt.Sprintf("%s/%s.js", rh.scriptDir, name))
	if err != nil {
		logger.Error("Failed to read file ", "error", err, "path", fmt.Sprintf("%s/%s.js", rh.scriptDir, name))
		return err
	}

	_, err = vm.RunString(string(b))

	if err != nil {
		return err
	}

	return nil
}

// renderActionRequest runs the action's Request JSON through the template
// engine with the report's funcMap and the URL query params, so request /
// filter / pipeline values can reference {{.profile_id}} etc. It mutates
// action.Request in place and is a no-op for an empty request. Shared by the
// NATS and mongo data-action paths.
func (rh ReportHandler) renderActionRequest(report *model.Report, action *model.DataActions, query map[string]any) error {

	if len(action.Request) == 0 {
		return nil
	}

	b, err := json.Marshal(action.Request)
	if err != nil {
		return err
	}

	t, err := template.New(report.Name).Funcs(rh.funcMap).Option("missingkey=zero").Parse(string(b))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, query); err != nil {
		return err
	}

	if err := json.Unmarshal(buf.Bytes(), &action.Request); err != nil {
		return err
	}

	if v, ok := action.Request["params"]; ok {
		if p, ok := v.(string); ok {
			action.Request["params"] = html.UnescapeString(p)
		}
	}

	return nil
}

func (rh ReportHandler) Generic(c echo.Context, reportType string, report *model.Report) error {

	templateName := fmt.Sprintf("%s.html", DefaultIfZero(report.TemplateName, "error"))

	query := make(map[string]any)

	for k, v := range c.QueryParams() {

		if len(v) > 1 {
			query[k] = v
		} else {

			query[k] = v[0]

			if k == "params" {
				query[k] = strings.ReplaceAll(query[k].(string), "\\", "")
			}
		}
	}

	data := make(map[string]any)

	for _, action := range report.DataActions {

		switch action.Type {

		case model.DATA_ACTION_JS:
			// Run a script with data and query; the script can mutate data.
			if err := rh.runScript(action.Name, data, query); err != nil {
				return c.Render(http.StatusOK, "error.html", err)
			}
			continue

		case model.DATA_ACTION_MONGO:
			// Query MongoDB directly for this dataset.
			if err := rh.renderActionRequest(report, action, query); err != nil {
				slog.Error("Failed to render mongo request", "action", action.Action, "error", err.Error())
				return c.Render(http.StatusOK, "error.html", err)
			}

			response, err := rh.conns.RunMongoDataAction(action)
			if err != nil {
				slog.Error("Failed to run mongo data action", "action", action.Action, "error", err.Error())
				return c.Render(http.StatusOK, "error.html", err)
			}

			slog.Info("Mongo response", "action", action.Name, "count", response["count"])
			data[action.Name] = response
			continue

		default:
			// NATS request/reply (type "nats" or unset).

			if len(action.TTL) == 0 {
				action.TTL = "120s"
			}

			ttl, err := time.ParseDuration(action.TTL)
			if err != nil {
				ttl = time.Second * 120
			}

			slog.Info("TTL is ", "ttl", ttl)

			action.Request["ttl"] = action.TTL

			if err := rh.renderActionRequest(report, action, query); err != nil {
				slog.Error("Failed to render request", "action", action.Action, "error", err.Error())
				return c.Render(http.StatusOK, "error.html", err)
			}

			slog.Info("Actions is ", "action", action)

			response, err := rh.serviceCall.Generic(action.Action, ttl, action.Request)
			if err != nil {
				slog.Error("Failed to run service", "action", action.Action, "error", err.Error())
				return c.Render(http.StatusOK, "error.html", err)
			}

			slog.Info("Response is", "response", response)

			data[action.Name] = response
		}
	}

	slog.Debug("Data is\n\n", "data", data)

	downLoadFileName := "BlueAsset_" + c.QueryParam("name")

	if reportType == "csv" {
		return rh.buildCSV(c, map[string]interface{}{"data": data, "query": query, "now": time.Now().Format("2006-01-02 15:04:05")}, downLoadFileName)
	} else {
		return c.Render(http.StatusOK, templateName, map[string]interface{}{"data": data, "query": query, "now": time.Now().Format("2006-01-02 15:04:05")})
	}

}

func DefaultIfZero[T comparable](value T, defaultValue T) T {
	var zero T // zero is automatically initialized to the zero value of type T
	if value == zero {
		return defaultValue
	}
	return value
}

func structToMap(item interface{}) map[string]interface{} {

	res := map[string]interface{}{}

	b, err := json.Marshal(item)
	if err != nil {
		slog.Error("Failed to marshall", "error", err)
		return res
	}

	err = json.Unmarshal(b, &res)
	if err != nil {
		slog.Error("Failed to unmarshall", "error", err)
		return res
	}

	return res
}

func buildFont(font *model.FontStyle) *excelize.Font {
	f := new(excelize.Font)

	if font.Bold {
		f.Bold = true
	}
	if font.Italic {
		f.Italic = true
	}

	if font.Family != "" {
		f.Family = font.Family
	}

	if font.Size != 0 {
		f.Size = font.Size
	}

	if font.Strike {
		f.Strike = true
	}

	if font.Color != "" {
		f.Color = font.Color
	}

	return f
}

func buildBorder(border []*model.BorderStyle) []excelize.Border {
	b := make([]excelize.Border, len(border))

	for index, border := range border {

		if border.Top {

			b[index].Type = "top"
			b[index].Color = border.Color
			b[index].Style = border.Style

		} else if border.Bottom {

			b[index].Type = "bottom"
			b[index].Color = border.Color
			b[index].Style = border.Style

		} else if border.Left {

			b[index].Type = "left"
			b[index].Color = border.Color
			b[index].Style = border.Style

		} else if border.Right {

			b[index].Type = "right"
			b[index].Color = border.Color
			b[index].Style = border.Style

		}

	}
	return b
}

func buildFill(fill *model.FillStyle) *excelize.Fill {
	f := new(excelize.Fill)

	if fill.Type != "" {
		f.Type = fill.Type
	}

	if fill.Color != nil {
		f.Color = fill.Color
	}

	if fill.Pattern != 0 {
		f.Pattern = fill.Pattern
	}

	if fill.Shading != 0 {
		f.Shading = fill.Shading
	}

	if fill.Transparency != 0 {
		f.Transparency = fill.Transparency
	}

	return f
}

func addCell(ef *excelize.File, colRow string, sheet string, cell *model.ExcelReporCell) {

	style := new(excelize.Style)

	if cell.Font != nil {
		style.Font = buildFont(cell.Font)
	}

	if cell.Fill != nil {
		style.Fill = *buildFill(cell.Fill)
	}

	if len(cell.Border) > 0 {
		style.Border = buildBorder(cell.Border)
	}

	_style, err := ef.NewStyle(style)
	if err == nil {
		ef.SetCellStyle(sheet, colRow, colRow, _style)
	}

	if cell.Type == "string" {
		v, err := cast.ToStringE(cell.Value)
		if err != nil {
			v = err.Error()
		}

		v = html.UnescapeString(v)

		// Set Value
		ef.SetCellValue(sheet, colRow, v)

	} else if cell.Type == "int" {
		v, err := cast.ToInt64E(cell.Value)
		if err != nil {
			slog.Error("Convert to int", "error", err)
			v = -999
		}
		style.NumFmt = 3
		ef.SetCellValue(sheet, colRow, v)

	} else if cell.Type == "float" {
		v, err := cast.ToFloat64E(cell.Value)
		if err != nil {
			slog.Error("Convert to float", "error", err)
			v = -999.99
		}
		style.NumFmt = 4
		ef.SetCellValue(sheet, colRow, v)

	} else if cell.Type == "date" {
		v, err := cast.StringToDate(cell.Value.(string))
		if err != nil {
			slog.Error("Convert to date", "error", err)
		}

		customDateFormat := "yyyy-m-dd hh:mm:ss"
		style.CustomNumFmt = &customDateFormat

		ef.SetCellValue(sheet, colRow, v)

	} else {
		ef.SetCellValue(sheet, colRow, cell.Value)
	}

}
