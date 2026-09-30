package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"blueassetgroup.com/reports-service/model"
	"blueassetgroup.com/reports-service/reportapi"
	utils "blueassetgroup.com/reports-service/shared"
	"github.com/labstack/echo/v4"
	"github.com/xuri/excelize/v2"

	_ "image/png"
)

type ReportHandler struct {
	csvRender   *reloadable[*template.Template]
	scripts     *ScriptRunner
	funcMap     template.FuncMap
	templateDir string
}

func NewReportHandler(config *utils.Config, funcMap template.FuncMap, scripts *ScriptRunner) *ReportHandler {
	csvRender := newTemplateReloadable(*config.TemplateDir, funcMap)
	return &ReportHandler{csvRender: csvRender, scripts: scripts, funcMap: funcMap, templateDir: *config.TemplateDir}
}

// xlsxContentType is the media type of the workbook buildCSV sends.
const xlsxContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

// buildCSV renders the report's Excel template — {{define "<template>.csv"}},
// or "<name>.csv" for a report without a template name — and sends the
// workbook it describes (see buildWorkbook).
func (self ReportHandler) buildCSV(c echo.Context, report *model.Report, data any, downLoadFileName string) error {
	csvTemplates, err := self.csvRender.Get()
	if err != nil {
		slog.Error("Error", "error", err)
		return renderError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to load csv templates %s", err.Error()))
	}

	name := report.TemplateName + ".csv"
	if report.TemplateName == "" || csvTemplates.Lookup(name) == nil {
		name = report.Name + ".csv"
	}

	var buf bytes.Buffer
	if err := csvTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		slog.Error("Error", "error", err)
		return renderError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to render %s", err.Error()))
	}

	excell := new(model.ExcelReport)
	if err := excell.FromJSON([]byte(strings.ReplaceAll(buf.String(), "\n", ""))); err != nil {
		slog.Error("The Excel template did not render valid JSON", "template", name, "error", err)
		return renderError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to read the Excel template %s: %s", name, err.Error()))
	}

	logo, _ := os.ReadFile(self.templateDir + "/logo.png")
	out, err := buildWorkbook(excell, report.Name, logo)
	if err != nil {
		slog.Error("Error", "error", err)
		return renderError(c, http.StatusInternalServerError, fmt.Sprintf("Failed to create excell sheet %s", err.Error()))
	}

	if excell.FileName != "" {
		downLoadFileName = excell.FileName
	}
	c.Response().Header().Set(echo.HeaderContentDisposition, mime.FormatMediaType("attachment", map[string]string{"filename": downLoadFileName + ".xlsx"}))
	return c.Blob(http.StatusOK, xlsxContentType, out.Bytes())
}

// flattenQuery splits echo's url.Values into a single-value map (last value
// wins) and a repeated-value map, for reportapi.Context.Query/QueryAll.
func flattenQuery(params map[string][]string) (map[string]string, map[string][]string) {
	query := make(map[string]string, len(params))
	queryAll := make(map[string][]string, len(params))

	for k, v := range params {
		queryAll[k] = v
		if len(v) == 0 {
			continue
		}
		value := v[len(v)-1]
		if k == "params" {
			value = strings.ReplaceAll(value, "\\", "")
		}
		query[k] = value
	}

	return query, queryAll
}

func (rh ReportHandler) Generic(c echo.Context, reportType string, report *model.Report) error {

	templateName := fmt.Sprintf("%s.html", DefaultIfZero(report.TemplateName, "error"))

	query, queryAll := flattenQuery(c.QueryParams())

	ctx := reportapi.Context{Query: query, QueryAll: queryAll, ReportName: report.Name, Now: time.Now()}

	data, err := rh.scripts.Run(report.Script, ctx)
	if err != nil {
		status := scriptErrorStatus(err)
		if status == http.StatusForbidden {
			slog.Info("Report not available to the caller", "script", report.Script, "reason", err.Error())
		} else {
			slog.Error("Failed to run report script", "script", report.Script, "error", err.Error())
		}
		return renderError(c, status, err)
	}

	slog.Debug("Data is\n\n", "data", data)

	downLoadFileName := "BlueAsset_" + c.QueryParam("name")

	if reportType == "csv" {
		return rh.buildCSV(c, report, map[string]interface{}{"data": data, "query": query, "now": ctx.Now.Format("2006-01-02 15:04:05")}, downLoadFileName)
	} else {
		return c.Render(http.StatusOK, templateName, map[string]interface{}{"data": data, "query": query, "now": ctx.Now.Format("2006-01-02 15:04:05")})
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
