package handlers

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"blueassetgroup.com/reports-service/repository"
	utils "blueassetgroup.com/reports-service/shared"
	globals "blueassetgroup.com/reports-service/utils"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"github.com/mafredri/cdp/devtool"
	"github.com/spf13/cast"
	"go.mongodb.org/mongo-driver/mongo"
)

// newServer wires the echo HTTP server: the html/csv template renderer, the
// static asset mount, the error handler and the /reports/run + /management
// routes. All report handlers are registered in the `globals` registry keyed by
// the report's Handler name ("generic" today). reportHandler is shared with
// the NATS-side Handler (see cmd/main.go) so lookup-parameter resolution
// reuses the same yaegi ScriptRunner as a full report run.
func NewServer(config *utils.Config, repo *repository.MongoRepository, funcMap template.FuncMap, reportHandler *ReportHandler) *echo.Echo {

	renderer := &templateRenderer{templates: newTemplateReloadable(*config.TemplateDir, funcMap)}
	// Fail fast on a broken template at startup, same as today's template.Must.
	if _, err := renderer.templates.Get(); err != nil {
		panic(err)
	}

	globals.AddHAndler("generic", reportHandler.Generic)

	e := echo.New()
	e.Renderer = renderer
	// static_dir is relative to the working directory. "../static" only worked
	// when the service was run from a subdirectory of the checkout; in the
	// container WORKDIR is /app and the assets are at /app/static, so a
	// hardcoded "../static" resolved to /static and every asset 404'd. Honour
	// the config value (which was declared but never read) and default to
	// "static" rather than "../static".
	staticDir := "static"
	if config.StaticDir != nil && *config.StaticDir != "" {
		staticDir = *config.StaticDir
	}
	e.Static("/reports/static", staticDir)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		slog.Error("Template render error", "error", err, "path", c.Request().URL.Path)
		c.Render(http.StatusInternalServerError, "error.html", err)
	}

	h := httpHandlers{config: config, repo: repo}
	e.GET("/run", h.runReport)
	e.GET("/management", h.management)

	return e
}

// httpHandlers holds the dependencies the HTTP routes need.
type httpHandlers struct {
	config *utils.Config
	repo   *repository.MongoRepository
}

// runReport serves GET /reports/run: it resolves the report by name, validates
// required params, then dispatches to the report handler for html / csv, or
// renders the html variant through headless Chrome for pdf.
func (h httpHandlers) runReport(c echo.Context) error {

	for key, value := range c.QueryParams() {
		slog.Info("Key/Value", "key", key, "value", value)
	}

	name := c.QueryParam("name")

	if len(name) == 0 {
		slog.Error("Name not optional")
		return renderError(c, http.StatusBadRequest, errors.New("name is not optional for now"))
	}

	_type := "html"

	if c.QueryParams().Has("type") {
		_type = c.QueryParam("type")
	}

	var err error

	report, err := h.repo.FindByName(name)

	if err != nil {
		slog.Error("Failed to find report", "error", err, "name", name)
		status := http.StatusInternalServerError
		if errors.Is(err, mongo.ErrNoDocuments) {
			status = http.StatusNotFound
		}
		return renderError(c, status, fmt.Errorf("Failed to find report with name %s, %w", name, err))

	}

	// Validate params
	for _, p := range report.Parameters {
		if p.Required && !c.QueryParams().Has(p.Name) {
			err = fmt.Errorf("Parameter %s is not optional", p.Name)
			slog.Error("Not optional", "error", err, "name", name)
			return renderError(c, http.StatusBadRequest, err)

		}

		if p.Regexp != "" && c.QueryParams().Has(p.Name) {
			re, reErr := regexp.Compile(p.Regexp)
			if reErr != nil {
				slog.Error("invalid parameter regexp", "name", p.Name, "regexp", p.Regexp, "error", reErr)
				continue
			}
			for _, v := range c.QueryParams()[p.Name] { // CHECKBOX sends repeated values
				if !re.MatchString(v) {
					err = fmt.Errorf("Parameter %s value %q does not match the required pattern", p.Name, v)
					slog.Error("Regexp mismatch", "error", err, "name", name)
					return renderError(c, http.StatusBadRequest, err)
				}
			}
		}
	}

	handlerName := report.Handler

	if len(handlerName) == 0 {
		handlerName = report.Name
	}

	// Run report
	if _type == "html" {

		handler, exist := globals.GetHandler(handlerName)
		if !exist {
			slog.Error("Failed to get handler", "error", "Handler does not exist for report", "name", handlerName)
			return renderError(c, http.StatusInternalServerError, fmt.Errorf("no handler %q for report %s", handlerName, name))
		}

		return handler(c, _type, report)

	} else if _type == "csv" {

		handler, exist := globals.GetHandler(handlerName)
		if !exist {
			slog.Error("Failed to get handler", "error", "Handler does not exist for report", "name", handlerName)
			return renderError(c, http.StatusInternalServerError, fmt.Errorf("no handler %q for report %s", handlerName, name))
		}

		return handler(c, _type, report)

	} else if _type == "pdf" {

		if h.config.ChromeUrl == nil {
			slog.Error("Chrome URL is not configured", "name", name)
			return renderError(c, http.StatusServiceUnavailable, "Chrome URL is not configured")
		}

		slog.Info("Chrome URL", "url", *h.config.ChromeUrl)

		// Bound the whole render; also aborts if the client disconnects.
		reqCtx, cancelReq := context.WithTimeout(c.Request().Context(), 60*time.Second)
		defer cancelReq()

		// Connect at the browser level so chromedp.NewContext opens a fresh tab
		// per request (and closes it on cancel) instead of sharing one page.
		ver, err := devtool.New(*h.config.ChromeUrl).Version(reqCtx)
		if err != nil {
			slog.Error("Failed to connect to chrome", "error", err, "name", name)
			return renderError(c, http.StatusServiceUnavailable, fmt.Sprintf("Failed to connect to chrome %s", err.Error()))
		}

		allocatorContext, cancelAlloc := chromedp.NewRemoteAllocator(reqCtx, ver.WebSocketDebuggerURL)
		defer cancelAlloc()

		ctx, cancelTab := chromedp.NewContext(allocatorContext)
		defer cancelTab()

		host := "localhost"
		if h.config.ChromeHost != nil {
			host = *h.config.ChromeHost
		}

		url := fmt.Sprintf("http://%s:%d%s?%s", host, h.config.Port, c.Path(), c.QueryString())
		slog.Info("URL is", "url", url)
		// capture pdf
		var buf []byte

		url = strings.ReplaceAll(url, "type=pdf", "type=html")

		slog.Info("URL for chrome", "url", url)

		// Load the HTML version first and look at how it answered: an error
		// page (a failed or refused report) must fail the PDF too, not be
		// printed as if it were the report.
		resp, err := chromedp.RunResponse(ctx, chromedp.Navigate(url))
		if err != nil {
			slog.Error("Error Running Chrome", "error", err)
			return renderError(c, http.StatusBadGateway, fmt.Sprintf("Failed to render %s", err.Error()))
		}
		if status, ok := pdfPageStatus(resp.Status); !ok {
			slog.Error("The report page failed, not printing it", "name", name, "page_status", resp.Status)
			return renderError(c, status, fmt.Sprintf("The report could not be produced (its page answered %d %s).", resp.Status, resp.StatusText))
		}
		if err := chromedp.Run(ctx, printToPDF(&buf)); err != nil {
			slog.Error("Error Running Chrome", "error", err)
			return renderError(c, http.StatusBadGateway, fmt.Sprintf("Failed to render %s", err.Error()))
		}

		c.Response().Header().Set(echo.HeaderContentDisposition,
			fmt.Sprintf("attachment; filename=%q", c.QueryParam("name")+".pdf"))
		err = c.Blob(http.StatusOK, "application/pdf", buf)
	} else {
		return renderError(c, http.StatusBadRequest, errors.New("type is not correctly set only options are csv | pdf | html"))
	}

	if err != nil {
		slog.Error("Failed to render", "error", err)
	}

	return err
}

// management serves GET /management.
func (h httpHandlers) management(c echo.Context) error {
	slog.Info("management called")
	return c.Render(200, "management.html", nil)
}

// templateRenderer adapts the parsed html templates to echo's Renderer.
type templateRenderer struct {
	templates *reloadable[*template.Template]
}

func (t *templateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, err := t.templates.Get()
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

// printToPDF captures the page the tab has loaded as a PDF into res.
func printToPDF(res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}

// fieldByName is the `fieldByName` template helper: reflective struct field
// access used by report templates.
func fieldByName(obj interface{}, fieldName string) (interface{}, error) {
	val := reflect.ValueOf(obj)
	// If the object is a pointer, dereference it
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("fieldByName: object is not a struct")
	}

	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("fieldByName: field '%s' not found", fieldName)
	}
	return field.Interface(), nil
}

// ReportFuncMap is the template function map shared by the html renderer, the
// csv renderer and the report data-action request templating.
func ReportFuncMap() template.FuncMap {
	return template.FuncMap{
		"formatAddress": func(address string) template.HTML {

			if strings.Contains(address, "@") {
				address = "<span class='font-bold'>" + strings.Split(address, "@")[0] + "</span><br><span>" + strings.Split(address, "@")[1] + "</span>"
			}

			return template.HTML(address)
		},
		"add": func(a, b any) float64 {

			_a := cast.ToFloat64(a)
			_b := cast.ToFloat64(b)
			//return math.Ceil((a + b) * 100 / 100)
			return _a + _b
		},
		"multiply": func(a, b int) int {
			return a * b
		},
		"formatFloat": func(a any) string {

			if i, ok := a.(int64); ok {
				return fmt.Sprintf("%d.00", i)
			} else if i, ok := a.(int); ok {
				return fmt.Sprintf("%d.00", i)
			}
			return fmt.Sprintf("%.2f", a)
		},
		"split": func(list string, sep string) []string {
			return strings.Split(list, sep)
		},
		"fieldByName": fieldByName,
		"convertToLocal": func(_t string) string {

			location, err := time.LoadLocation("Africa/Johannesburg")

			if err != nil {
				return err.Error()
			}

			date, err := time.Parse(time.RFC3339, _t)

			if err != nil {
				return err.Error()
			}

			_date := date.In(location)

			return _date.Format("2006-01-02 15:04:05")
		},
		"formatWorkingTime": func(a any) string {

			totalMinutes := cast.ToInt(a)

			hours := totalMinutes / 60
			minutes := totalMinutes % 60
			return fmt.Sprintf("%d:%02d", hours, minutes)

		}, "comma": func(f any) string {
			_f := cast.ToInt64(f)

			return humanize.Comma(_f)
		}, "mod": func(a, b int) int {
			return a % b
		}, "replaceAll": strings.ReplaceAll,
		"fixString": func(s string) string {
			return strings.ReplaceAll(strings.ReplaceAll(s, "\\N", ""), "\t", "")
		},
		"money": formatMoney,
		"pct":   formatPct,
	}
}

// formatMoney shows cents in a currency: "R 1 234.56" for ZAR (or no
// currency), "USD 1 234.56" otherwise, with a space every three digits —
// the broker portal's own style. Money in the portal's collections is whole
// cents (models.Money), so cents is anything that casts to an int64.
func formatMoney(cents any, currency string) string {
	n := cast.ToInt64(cents)
	sign := ""
	if n < 0 {
		sign, n = "-", -n
	}
	digits := strconv.FormatInt(n/100, 10)
	var b strings.Builder
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(d)
	}
	return fmt.Sprintf("%s %s%s.%02d", currencySymbol(currency), sign, b.String(), n%100)
}

// formatPct shows a percentage without trailing zeros: "2.1%", "70%".
func formatPct(v any) string {
	return strconv.FormatFloat(cast.ToFloat64(v), 'f', -1, 64) + "%"
}
