//go:build ignore

// demo_all_param_types.go is a report script: real Go syntax interpreted by
// yaegi at request time (see handlers/scriptrunner.go), not compiled into
// this module's binary — the ignore tag above keeps `go build ./...`/`go vet
// ./...` from trying to build every script file as one Go package (they all
// declare `package script` and would otherwise collide).
//
// demo_all_param_types.go — echoes back whatever was submitted on the
// "demo-all-param-types" report (see
// mongo/demo-all-param-types-report.mongodb.js), so the run screen just
// shows exactly what each parameter type actually sent through. all_values
// carries every value for a field that can repeat (the "tags" CHECKBOX
// group); the html/http.Generic caller already passes the single-value
// equivalent (ctx.Query, last value wins) to the template as `.query`.
package script

import "blueassetgroup.com/reports-service/reportapi"

func Run(ctx reportapi.Context) (map[string]any, error) {
	return map[string]any{
		"all_values": ctx.QueryAll,
	}, nil
}
