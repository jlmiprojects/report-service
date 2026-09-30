package handlers

import (
	"errors"
	"net/http"

	"blueassetgroup.com/reports-service/reportapi"
	"github.com/labstack/echo/v4"
)

// renderError answers a /run request that produced no report: the error page,
// with a real status. It used to be sent with 200, and for type=pdf Chrome
// then printed the error page as if it were the report — so a caller such as
// the broker portal, which can only look at the status and content type,
// took a failure for a download.
//
//	400  missing name, parameter or type; a parameter that fails its pattern
//	403  the script found nothing the caller may see (reportapi.NotAvailable)
//	404  no report by that name
//	500  the script, template or workbook failed
//	502  Chrome could not render or print the page
//	503  Chrome is not configured or can't be reached
func renderError(c echo.Context, status int, msg any) error {
	return c.Render(status, "error.html", msg)
}

// scriptErrorStatus is the status for an error a report script returned:
// 403 when it said the report isn't available to the caller, else 500.
func scriptErrorStatus(err error) int {
	var na *reportapi.NotAvailableError
	if errors.As(err, &na) {
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}

// pdfPageStatus decides what a PDF request answers once Chrome has loaded the
// report's HTML page: ok when the page answered 200, so it may be printed;
// otherwise the status to fail with. A caller's error (400, 403, 404) is
// passed on as it is; anything else means the report failed (502).
func pdfPageStatus(pageStatus int64) (status int, ok bool) {
	switch pageStatus {
	case http.StatusOK:
		return http.StatusOK, true
	case http.StatusBadRequest, http.StatusForbidden, http.StatusNotFound:
		return int(pageStatus), false
	default:
		return http.StatusBadGateway, false
	}
}
