package handlers

import (
	"testing"

	"blueassetgroup.com/reports-service/model"
	"github.com/xuri/excelize/v2"
)

// Every fixed bug in one workbook: more than 26 columns (column 15 was
// written as "0" and column 27 crashed), number formats that apply, a date
// given as a string, money in its own currency, and a bold font.
func TestBuildWorkbook(t *testing.T) {
	var wide []model.ExcelReporCell
	for i := range 30 {
		wide = append(wide, model.ExcelReporCell{Value: i + 1, Type: "int"})
	}
	report := &model.ExcelReport{
		FreezeRows: 1,
		FileName:   "statement",
		Rows: [][]model.ExcelReporCell{
			{{Value: "Name &amp; co", Type: "string", Font: &model.FontStyle{Bold: true, Color: "#FFFFFF"}}},
			wide,
			{
				{Value: 123456, Type: "money", Currency: "ZAR"},
				{Value: -5000, Type: "money", Currency: "usd"},
				{Value: "2025-09-25", Type: "date"},
				{Value: 2.5, Type: "float"},
			},
		},
	}

	buf, err := buildWorkbook(report, "commission-statement: a very long report name", nil)
	if err != nil {
		t.Fatal(err)
	}
	f, err := excelize.OpenReader(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sheet := f.GetSheetName(0)
	if len([]rune(sheet)) > 31 {
		t.Errorf("sheet name %q is longer than Excel allows", sheet)
	}

	if v, _ := f.GetCellValue(sheet, "A1"); v != "Name & co" {
		t.Errorf("A1 = %q", v)
	}
	if v, _ := f.GetCellValue(sheet, "O2", excelize.Options{RawCellValue: true}); v != "15" {
		t.Errorf("O2 (column 15) = %q, want 15", v)
	}
	if v, _ := f.GetCellValue(sheet, "AD2", excelize.Options{RawCellValue: true}); v != "30" {
		t.Errorf("AD2 (column 30) = %q, want 30", v)
	}

	// Money is a number in rands, formatted in its currency.
	if v, _ := f.GetCellValue(sheet, "A3", excelize.Options{RawCellValue: true}); v != "1234.56" {
		t.Errorf("A3 raw = %q, want 1234.56", v)
	}
	if v, _ := f.GetCellValue(sheet, "A3"); v != "R 1,234.56" {
		t.Errorf("A3 shown = %q, want R 1,234.56", v)
	}
	if v, _ := f.GetCellValue(sheet, "B3"); v != "-USD 50.00" {
		t.Errorf("B3 shown = %q, want -USD 50.00", v)
	}
	if v, _ := f.GetCellValue(sheet, "C3"); v != "2025-09-25" {
		t.Errorf("C3 shown = %q, want 2025-09-25", v)
	}
	if v, _ := f.GetCellValue(sheet, "D3"); v != "2.50" {
		t.Errorf("D3 shown = %q, want 2.50", v)
	}

	id, _ := f.GetCellStyle(sheet, "A1")
	style, _ := f.GetStyle(id)
	if style == nil || style.Font == nil || !style.Font.Bold {
		t.Errorf("A1 is not bold: %+v", style)
	}
}

func TestBuildWorkbookRejectsBadCells(t *testing.T) {
	report := &model.ExcelReport{Rows: [][]model.ExcelReporCell{{{Value: "twelve", Type: "money"}}}}
	if _, err := buildWorkbook(report, "x", nil); err == nil {
		t.Error("a non-numeric money cell was accepted")
	}
}

func TestFormatMoney(t *testing.T) {
	cases := []struct {
		cents    any
		currency string
		want     string
	}{
		{int64(123456789), "ZAR", "R 1 234 567.89"},
		{420000, "", "R 4 200.00"},
		{int64(-5677), "USD", "USD -56.77"},
		{"795", "ZAR", "R 7.95"},
	}
	for _, c := range cases {
		if got := formatMoney(c.cents, c.currency); got != c.want {
			t.Errorf("formatMoney(%v, %q) = %q, want %q", c.cents, c.currency, got, c.want)
		}
	}
	if got := formatPct(2.1); got != "2.1%" {
		t.Errorf("formatPct = %q", got)
	}
}
