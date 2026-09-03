package utils

import (
	"blueassetgroup.com/reports-service/model"
	"github.com/labstack/echo/v4"
)

var Handlers map[string]func(c echo.Context, reportType string, report *model.Report) error

func AddHAndler(name string, handler func(c echo.Context, reportType string, report *model.Report) error) {
	if Handlers == nil {
		Handlers = make(map[string]func(c echo.Context, reportType string, report *model.Report) error)
	}
	Handlers[name] = handler
}

func GetHandler(name string) (func(c echo.Context, reportType string, report *model.Report) error, bool) {
	handler, ok := Handlers[name]
	if !ok {
		return nil, false
	}
	return handler, true
}
