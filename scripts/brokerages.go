//go:build ignore

// brokerages.go is a report script: real Go syntax interpreted by yaegi at
// request time (see handlers/scriptrunner.go), not compiled into this
// module's binary — the ignore tag above keeps `go build ./...`/`go vet
// ./...` from trying to build every script file as one Go package (they all
// declare `package script` and would otherwise collide).
//
// brokerages.go — example LOOKUP options script: lists brokerages from the
// broker_portal database as {name,value} options for a "lookup"-typed report
// parameter (see model.LookupMetadata). Optionally scoped by a "country"
// value already chosen elsewhere on the same run form — reportapi.Context.
// Query carries every currently-submitted parameter value (not just this
// parameter's own), so a LOOKUP script can cascade off a sibling parameter.
// Brokerage documents store the country as a nested "address.country_code"
// field (broker-portal/ui/models/{brokerage,utils}.go), not a top-level
// "country" field.
package script

import (
	"fmt"
	"log/slog"

	"blueassetgroup.com/reports-service/reportapi"
)

func Options(ctx reportapi.Context) ([]reportapi.Option, error) {

	slog.Info("Options is ","ctx",ctx)
	filter := map[string]any{}
	if country := ctx.Query["country"]; country != "" {
		filter["address.country_code"] = country
	}

	res, err := ctx.Mongo("broker_portal").Find("", "brokerages", filter, map[string]any{
		"projection": map[string]any{"name": 1},
		"sort":       map[string]any{"name": 1},
		"limit":      200,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve brokerages: %w", err)
	}

	options := make([]reportapi.Option, 0, len(res.Data))
	for _, row := range res.Data {
		id, _ := row["_id"].(string)
		name, _ := row["name"].(string)
		if id == "" || name == "" {
			continue
		}
		options = append(options, reportapi.Option{Name: name, Value: id})
	}
	return options, nil
}
