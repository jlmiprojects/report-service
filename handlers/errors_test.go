package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"blueassetgroup.com/reports-service/model"
	"blueassetgroup.com/reports-service/reportapi"
	utils "blueassetgroup.com/reports-service/shared"
	"github.com/labstack/echo/v4"
)

func TestPdfPageStatus(t *testing.T) {
	cases := []struct {
		page   int64
		status int
		ok     bool
	}{
		{200, 200, true},
		{400, 400, false},
		{403, 403, false},
		{404, 404, false},
		{500, 502, false},
		{503, 502, false},
	}
	for _, c := range cases {
		if status, ok := pdfPageStatus(c.page); status != c.status || ok != c.ok {
			t.Errorf("pdfPageStatus(%d) = %d, %v; want %d, %v", c.page, status, ok, c.status, c.ok)
		}
	}
}

// A script's reportapi.NotAvailable is created inside the interpreter; it must
// still be the host's *reportapi.NotAvailableError for errors.As to see it.
func TestNotAvailableCrossesTheInterpreter(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, filepath.Join(dir, "refuse.go"), `package script

import "blueassetgroup.com/reports-service/reportapi"

func Run(ctx reportapi.Context) (map[string]any, error) {
	return nil, reportapi.NotAvailable("This statement isn't available to you.")
}
`)
	writeScript(t, filepath.Join(dir, "broken.go"), `package script

import (
	"fmt"

	"blueassetgroup.com/reports-service/reportapi"
)

func Run(ctx reportapi.Context) (map[string]any, error) {
	return nil, fmt.Errorf("load statement: connection refused")
}
`)
	sr := NewScriptRunner(dir)

	_, err := sr.Run("refuse", reportapi.Context{})
	var na *reportapi.NotAvailableError
	if !errors.As(err, &na) || na.Reason != "This statement isn't available to you." {
		t.Fatalf("refusal came back as %T %v", err, err)
	}
	if got := scriptErrorStatus(err); got != http.StatusForbidden {
		t.Errorf("refusal status = %d, want 403", got)
	}

	_, err = sr.Run("broken", reportapi.Context{})
	if err == nil || scriptErrorStatus(err) != http.StatusInternalServerError {
		t.Errorf("failure: %v, status %d, want 500", err, scriptErrorStatus(err))
	}
}

// Generic answers a refused report with 403 and a failed one with 500, in
// every format — never 200, which a caller would take for the report.
func TestGenericErrorStatuses(t *testing.T) {
	dir := t.TempDir()
	writeScript(t, filepath.Join(dir, "refuse.go"), `package script

import "blueassetgroup.com/reports-service/reportapi"

func Run(ctx reportapi.Context) (map[string]any, error) {
	return nil, reportapi.NotAvailable("Not yours.")
}
`)
	writeScript(t, filepath.Join(dir, "broken.go"), `package script

import (
	"fmt"

	"blueassetgroup.com/reports-service/reportapi"
)

func Run(ctx reportapi.Context) (map[string]any, error) {
	return nil, fmt.Errorf("boom")
}
`)

	templateDir := filepath.Join("..", "reports", "templates")
	funcs := ReportFuncMap()
	e := echo.New()
	e.Renderer = &templateRenderer{templates: newTemplateReloadable(templateDir, funcs)}
	rh := NewReportHandler(&utils.Config{TemplateDir: &templateDir}, funcs, NewScriptRunner(dir))

	cases := []struct {
		script, format string
		status         int
		body           string
	}{
		{"refuse", "html", http.StatusForbidden, "Not yours."},
		{"refuse", "csv", http.StatusForbidden, "Not yours."},
		{"broken", "html", http.StatusInternalServerError, "boom"},
		{"broken", "csv", http.StatusInternalServerError, "boom"},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		ctx := e.NewContext(httptest.NewRequest(http.MethodGet, "/run?name=x&type="+c.format, nil), rec)
		if err := rh.Generic(ctx, c.format, &model.Report{Name: "x", Script: c.script, TemplateName: "x"}); err != nil {
			t.Fatalf("%s/%s: %v", c.script, c.format, err)
		}
		if rec.Code != c.status || !strings.Contains(rec.Body.String(), c.body) {
			t.Errorf("%s/%s: status %d (want %d), body has %q: %v", c.script, c.format, rec.Code, c.status, c.body, strings.Contains(rec.Body.String(), c.body))
		}
	}
}
