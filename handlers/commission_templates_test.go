package handlers

import (
	"bytes"
	"html/template"
	"path/filepath"
	"strings"
	"testing"

	"blueassetgroup.com/reports-service/model"
	"github.com/xuri/excelize/v2"
)

// The commission statement templates live in the reports repo (through the
// reports -> ../reports symlink). This renders each one — PDF (html) and
// Excel — from data shaped like its script's output, and reads the Excel
// JSON back into a workbook, so a template that no longer renders, or renders
// invalid JSON, fails here rather than in front of a brokerage.

func commissionTemplates(t *testing.T) *template.Template {
	t.Helper()
	tpl, err := template.New("").Funcs(ReportFuncMap()).ParseGlob(filepath.Join("..", "reports", "templates", "*.html"))
	if err != nil {
		t.Fatalf("parse templates: %v", err)
	}
	return tpl
}

func renderBoth(t *testing.T, tpl *template.Template, name string, data map[string]any) (string, *excelize.File) {
	t.Helper()
	in := map[string]any{"data": data, "query": map[string]string{}, "now": "2025-10-25 10:00:00"}

	var html bytes.Buffer
	if err := tpl.ExecuteTemplate(&html, name+".html", in); err != nil {
		t.Fatalf("%s.html: %v", name, err)
	}

	var js bytes.Buffer
	if err := tpl.ExecuteTemplate(&js, name+".csv", in); err != nil {
		t.Fatalf("%s.csv: %v", name, err)
	}
	report := new(model.ExcelReport)
	if err := report.FromJSON([]byte(strings.ReplaceAll(js.String(), "\n", ""))); err != nil {
		t.Fatalf("%s.csv renders invalid JSON: %v\n%s", name, err, js.String())
	}
	buf, err := buildWorkbook(report, name, nil)
	if err != nil {
		t.Fatalf("%s workbook: %v", name, err)
	}
	f, err := excelize.OpenReader(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return html.String(), f
}

// findRow returns the cells of the first row whose first cell is label.
func findRow(t *testing.T, f *excelize.File, label string) []string {
	t.Helper()
	rows, err := f.GetRows(f.GetSheetName(0), excelize.Options{RawCellValue: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rows {
		if len(r) > 0 && r[0] == label {
			return r
		}
	}
	t.Fatalf("no row %q", label)
	return nil
}

var sampleEntries = []map[string]any{{
	"name": "Thandi O'Brien", "total": int64(420000),
	"entries": []map[string]any{{
		"plan_id": "plan-1", "kind": "Once-off", "period": "—", "recipient": "Sipho \"Sy\" Mokoena", "recipient_type": "Advisor",
		"base": int64(20000000), "rate_pct": 2.1, "share_pct": 80.0, "amount": int64(235200), "note": "",
	}},
}}

func TestCommissionStatementTemplates(t *testing.T) {
	tpl := commissionTemplates(t)
	kinds := map[string]any{"once_off": int64(420000), "recurring": int64(0), "adjustment": int64(0)}
	data := map[string]any{
		"statement": map[string]any{
			"brokerage_name": "Mokoena Wealth", "statement_run": "2025-09", "recurring_period": "2025-08",
			"currency": "ZAR", "run_date": "2025-09-25", "vat_number": "4123456789", "vat_rate_pct": 15.0,
			"amount": int64(420000), "vat": int64(63000), "amount_incl_vat": int64(483000),
			"kinds":  kinds,
			"totals": map[string]any{"brokerage": int64(126000), "advisors": int64(294000)},
			"lines": []any{
				map[string]any{"type": "BROKERAGE", "name": "Mokoena Wealth", "amount": int64(126000), "own": int64(0), "kinds": kinds},
				map[string]any{"type": "ADVISOR", "name": "Sipho Mokoena", "amount": int64(294000), "own": int64(235200), "kinds": kinds,
					"owes": []any{map[string]any{"name": "Lindiwe Dube", "amount": int64(58800)}}},
			},
		},
		"currency": "ZAR", "run_date": "25 Sep 2025", "issuer_line": "BluHive · Reg. 2020/123", "vat_label": "VAT 15% on the full amount",
		"clients": sampleEntries, "entry_count": 1, "generated": "25 Oct 2025 10:00",
	}

	html, f := renderBoth(t, tpl, "commission_statement", data)
	for _, want := range []string{"R 4 200.00", "R 630.00", "R 4 830.00", "owes referrer Lindiwe Dube", "R 588.00", "Thandi O&#39;Brien"} {
		if !strings.Contains(html, want) {
			t.Errorf("statement PDF lacks %q", want)
		}
	}
	if r := findRow(t, f, "Total including VAT"); len(r) < 2 || r[1] != "4830" {
		t.Errorf("Excel total incl. VAT row = %v, want 4830", r)
	}
	if r := findRow(t, f, "Thandi O'Brien"); len(r) < 9 || r[4] != `Sipho "Sy" Mokoena (Advisor)` || r[8] != "2352" {
		t.Errorf("Excel entry row = %v", r)
	}

	// A statement the viewer may not see isn't rendered at all: its script
	// returns reportapi.NotAvailable, answered with 403 (errors_test.go).
}

func TestCommissionStatementRangeTemplates(t *testing.T) {
	tpl := commissionTemplates(t)
	data := map[string]any{
		"brokerage_name": "Mokoena Wealth", "from": "2025-03", "to": "2026-02", "generated": "now",
		"runs": []map[string]any{
			{"run": "2025-09", "run_date": "25 Sep 2025", "currency": "ZAR", "amount": int64(420000), "vat": int64(0), "amount_incl_vat": int64(420000)},
			{"run": "2025-10", "run_date": "25 Oct 2025", "currency": "USD", "amount": int64(5000), "vat": int64(0), "amount_incl_vat": int64(5000)},
		},
		"totals": []map[string]any{
			{"currency": "USD", "once_off": int64(5000), "recurring": int64(0), "adjustment": int64(0), "amount": int64(5000), "vat": int64(0), "amount_incl_vat": int64(5000), "brokerage": int64(1500), "advisors": int64(3500)},
			{"currency": "ZAR", "once_off": int64(420000), "recurring": int64(0), "adjustment": int64(0), "amount": int64(420000), "vat": int64(0), "amount_incl_vat": int64(420000), "brokerage": int64(126000), "advisors": int64(294000)},
		},
		"advisors": []map[string]any{{"currency": "ZAR", "name": "Sipho Mokoena", "amount": int64(294000), "own": int64(235200), "owed": int64(58800)}},
	}
	html, f := renderBoth(t, tpl, "commission_statement_range", data)
	for _, want := range []string{"USD 50.00", "R 4 200.00", "Sipho Mokoena", "2025-03 to 2026-02"} {
		if !strings.Contains(html, want) {
			t.Errorf("range PDF lacks %q", want)
		}
	}
	if r := findRow(t, f, "USD"); len(r) < 7 || r[6] != "50" {
		t.Errorf("Excel USD totals row = %v", r)
	}
}

func TestCommissionAdvisorStatementTemplates(t *testing.T) {
	tpl := commissionTemplates(t)
	data := map[string]any{
		"advisor_name": "Sipho Mokoena", "run": "2025-09", "currency": "ZAR", "generated": "now",
		"share": int64(294000), "referrers": int64(58800), "own": int64(235200),
		"owes":    []map[string]any{{"name": "Lindiwe Dube", "amount": int64(58800)}},
		"clients": sampleEntries, "entry_count": 1,
	}
	html, f := renderBoth(t, tpl, "commission_advisor_statement", data)
	for _, want := range []string{"R 2 940.00", "R 588.00", "R 2 352.00", "Lindiwe Dube"} {
		if !strings.Contains(html, want) {
			t.Errorf("advisor PDF lacks %q", want)
		}
	}
	if r := findRow(t, f, "Yours after referrers"); len(r) < 2 || r[1] != "2352" {
		t.Errorf("Excel 'yours' row = %v", r)
	}
}

// The plan report renders a plan's movements and entries, in both formats,
// including a correction, and with nothing in the range.
func TestPlanCommissionTemplates(t *testing.T) {
	tpl := commissionTemplates(t)
	full := map[string]any{
		"plan_id": "sim-plan-0001", "label": "Q-SIM-0001", "advisor": "Sipho Mokoena", "currency": "ZAR",
		"from": "1 Sep 2025", "to": "31 Dec 2025", "run_from": "2025-09", "run_to": "2025-12",
		"opening": int64(0), "loans": int64(20000000), "profits": int64(160000), "corrections": int64(-80000), "closing": int64(20080000),
		"movements": []map[string]any{
			{"date": "15 Sep 2025", "kind": "Loan", "amount": int64(20000000), "value": int64(20000000), "note": ""},
			{"date": "16 Sep 2025", "kind": "Profit", "amount": int64(80000), "value": int64(20080000), "note": ""},
			{"date": "17 Sep 2025", "kind": "Correction", "amount": int64(-80000), "value": int64(20000000), "note": "posted twice"},
		},
		"entries": []map[string]any{
			{"run": "2025-09", "kind": "Once-off", "period": "", "recipient": "Mokoena Wealth", "recipient_type": "Brokerage",
				"base": int64(20000000), "rate": 2.1, "amount": int64(126000), "currency": "ZAR"},
		},
		"totals":    []map[string]any{{"currency": "ZAR", "amount": int64(126000)}},
		"generated": "6 Oct 2026 13:35",
	}
	html, _ := renderBoth(t, tpl, "plan_commission", full)
	if !strings.Contains(html, "posted twice") || !strings.Contains(html, "Corrections") {
		t.Fatal("the html is missing the correction")
	}

	empty := map[string]any{}
	for k, v := range full {
		empty[k] = v
	}
	empty["movements"], empty["entries"], empty["totals"] = []map[string]any{}, []map[string]any{}, []map[string]any{}
	renderBoth(t, tpl, "plan_commission", empty)
}
