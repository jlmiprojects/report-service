package handlers

import (
	"bytes"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"blueassetgroup.com/reports-service/model"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

// buildWorkbook turns a rendered Excel template (model.ExcelReport) into an
// .xlsx: one sheet named sheet, the template's rows cell by cell, and its
// header, footer and optional logo.
//
// Each cell's style — font, fill, border and number format — is complete
// before it is registered, so every part of it applies. Cell types:
//
//	string  text (HTML entities from the template are unescaped)
//	int     whole number, #,##0
//	float   number, #,##0.00
//	date    a date (string in any common layout, or a time), yyyy-mm-dd
//	money   cents as an integer, written as a number in the cell's currency:
//	        "R" #,##0.00 for ZAR (or no currency), "USD" #,##0.00 otherwise
//
// A cell's Format, when set, replaces the type's number format.
func buildWorkbook(report *model.ExcelReport, sheet string, logo []byte) (*bytes.Buffer, error) {
	sheet = excelSheetName(sheet)

	ef := excelize.NewFile()
	defer ef.Close()

	if err := ef.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}

	fitToPage := false
	margin, leftMargin := 1.5, 0.5
	_ = ef.SetSheetProps(sheet, &excelize.SheetPropsOptions{FitToPage: &fitToPage})
	_ = ef.SetPageMargins(sheet, &excelize.PageLayoutMarginsOptions{Top: &margin, Bottom: &margin, Left: &leftMargin})
	if o := report.PageLayout.Orientation; o == "landscape" || o == "portrait" {
		_ = ef.SetPageLayout(sheet, &excelize.PageLayoutOptions{Orientation: &o})
	}

	if len(logo) > 0 {
		if err := ef.AddHeaderFooterImage(sheet, &excelize.HeaderFooterImageOptions{
			Position: excelize.HeaderFooterImagePositionRight, Extension: ".png",
			Width: "120pt", Height: "30pt", File: logo,
		}); err != nil {
			slog.Error("Failed to add the header logo", "error", err)
		}
	}
	noScale := false
	if err := ef.SetHeaderFooter(sheet, &excelize.HeaderFooterOptions{
		OddHeader: report.Header, OddFooter: report.Footer,
		AlignWithMargins: &noScale, ScaleWithDoc: &noScale,
	}); err != nil {
		slog.Error("Failed to add the header and footer", "error", err)
	}

	styles := map[string]int{}
	widths := map[int]float64{}
	for r, row := range report.Rows {
		for c := range row {
			ref, err := excelize.CoordinatesToCellName(c+1, r+1)
			if err != nil {
				return nil, err
			}
			shown, err := writeCell(ef, sheet, ref, &row[c], styles)
			if err != nil {
				return nil, fmt.Errorf("cell %s: %w", ref, err)
			}
			if w := float64(utf8.RuneCountInString(shown)) + 2; w > widths[c] {
				widths[c] = w
			}
		}
	}

	for c, w := range widths {
		if c < len(report.ColumnWidths) && report.ColumnWidths[c] > 0 {
			w = report.ColumnWidths[c]
		}
		name, _ := excelize.ColumnNumberToName(c + 1)
		_ = ef.SetColWidth(sheet, name, name, min(w, 60))
	}

	if report.FreezeRows > 0 {
		top, _ := excelize.CoordinatesToCellName(1, report.FreezeRows+1)
		_ = ef.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: report.FreezeRows, TopLeftCell: top, ActivePane: "bottomLeft"})
	}

	return ef.WriteToBuffer()
}

// writeCell writes one cell with its complete style, and returns the text it
// shows, for sizing the column.
func writeCell(ef *excelize.File, sheet, ref string, cell *model.ExcelReporCell, styles map[string]int) (string, error) {
	style := &excelize.Style{}
	if cell.Font != nil {
		style.Font = buildFont(cell.Font)
	}
	if cell.Fill != nil {
		style.Fill = *buildFill(cell.Fill)
	}
	if len(cell.Border) > 0 {
		style.Border = buildBorder(cell.Border)
	}

	var value any
	var shown string
	numFmt := ""
	switch cell.Type {
	case "string", "":
		s := html.UnescapeString(cast.ToString(cell.Value))
		value, shown = s, s
	case "int":
		v, err := cast.ToInt64E(cell.Value)
		if err != nil {
			return "", fmt.Errorf("not a whole number: %v", cell.Value)
		}
		value, shown, numFmt = v, fmt.Sprint(v), "#,##0"
	case "float":
		v, err := cast.ToFloat64E(cell.Value)
		if err != nil {
			return "", fmt.Errorf("not a number: %v", cell.Value)
		}
		value, shown, numFmt = v, fmt.Sprintf("%.2f", v), "#,##0.00"
	case "money":
		cents, err := cast.ToInt64E(cell.Value)
		if err != nil {
			return "", fmt.Errorf("money is whole cents, got %v", cell.Value)
		}
		symbol := currencySymbol(cell.Currency)
		value = float64(cents) / 100
		shown = symbol + " " + fmt.Sprintf("%.2f", value)
		numFmt = fmt.Sprintf(`"%s" #,##0.00;-"%s" #,##0.00`, symbol, symbol)
	case "date":
		t, err := cellTime(cell.Value)
		if err != nil {
			return "", err
		}
		value, shown, numFmt = t, t.Format("2006-01-02"), "yyyy-mm-dd"
	default:
		value, shown = cell.Value, cast.ToString(cell.Value)
	}
	if cell.Format != "" {
		numFmt = cell.Format
	}
	if numFmt != "" {
		f := numFmt
		style.CustomNumFmt = &f
	}

	id, err := styleID(ef, style, styles)
	if err != nil {
		return "", err
	}
	if err := ef.SetCellValue(sheet, ref, value); err != nil {
		return "", err
	}
	return shown, ef.SetCellStyle(sheet, ref, ref, id)
}

// styleID registers style once and reuses it for every identical cell.
func styleID(ef *excelize.File, style *excelize.Style, cache map[string]int) (int, error) {
	key := fmt.Sprintf("%+v|%+v|%+v|%v", style.Font, style.Fill, style.Border, deref(style.CustomNumFmt))
	if id, ok := cache[key]; ok {
		return id, nil
	}
	id, err := ef.NewStyle(style)
	if err != nil {
		return 0, err
	}
	cache[key] = id
	return id, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func cellTime(v any) (time.Time, error) {
	if t, ok := v.(time.Time); ok {
		return t, nil
	}
	t, err := cast.StringToDate(cast.ToString(v))
	if err != nil {
		return time.Time{}, fmt.Errorf("not a date: %v", v)
	}
	return t, nil
}

// currencySymbol is how an amount in currency is shown: "R" for Rand (or no
// currency), the ISO code otherwise.
func currencySymbol(currency string) string {
	if currency == "" || strings.EqualFold(currency, "ZAR") {
		return "R"
	}
	return strings.ToUpper(currency)
}

// excelSheetName makes name safe as a sheet name: at most 31 characters and
// none of []:*?/\.
func excelSheetName(name string) string {
	name = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`[]:*?/\`, r) {
			return '-'
		}
		return r
	}, strings.TrimSpace(name))
	if name == "" {
		name = "Report"
	}
	if utf8.RuneCountInString(name) > 31 {
		name = string([]rune(name)[:31])
	}
	return name
}

func buildFont(font *model.FontStyle) *excelize.Font {
	return &excelize.Font{
		Bold:      font.Bold,
		Italic:    font.Italic,
		Underline: font.Underline,
		Family:    font.Family,
		Size:      font.Size,
		Strike:    font.Strike,
		Color:     strings.TrimPrefix(font.Color, "#"),
	}
}
