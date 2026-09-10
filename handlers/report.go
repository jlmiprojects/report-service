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
	"blueassetgroup.com/reports-service/reportapi"
	utils "blueassetgroup.com/reports-service/shared"
	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"
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

	csvTemplates, err := self.csvRender.Get()
	if err != nil {
		slog.Error("Error", "error", err)
		return c.Render(200, "error.html", fmt.Sprintf("Failed to load csv templates %s", err.Error()))
	}

	var buf bytes.Buffer

	err = csvTemplates.ExecuteTemplate(&buf, c.QueryParam("name")+".csv", data)

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
		slog.Error("Failed to run report script", "script", report.Script, "error", err.Error())
		return c.Render(http.StatusOK, "error.html", err)
	}

	slog.Debug("Data is\n\n", "data", data)

	downLoadFileName := "BlueAsset_" + c.QueryParam("name")

	if reportType == "csv" {
		return rh.buildCSV(c, map[string]interface{}{"data": data, "query": query, "now": ctx.Now.Format("2006-01-02 15:04:05")}, downLoadFileName)
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
