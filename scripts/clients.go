//go:build ignore

// clients.go is a report script: real Go syntax interpreted by yaegi at
// request time (see handlers/scriptrunner.go), not compiled into this
// module's binary — the ignore tag above keeps `go build ./...`/`go vet
// ./...` from trying to build every script file as one Go package (they all
// declare `package script` and would otherwise collide).
//
// clients.go — clients listing, optionally scoped to one advisor or a
// principal's book. See conf/clients_report.json ("script": "clients").
// GET /run?name=clients[&advisor_id=<hex>][&principal_id=<hex>][&type=csv].
//
// advisor_id (optional) scopes to one advisor's clients; principal_id
// (optional, not a declared parameter — broker-portal injects it as a hidden
// field alongside an "All" advisor_id selection) scopes to every client
// under advisors reporting to that principal. Neither set = every client
// (ADMIN "All").
package script

import (
	"fmt"

	"blueassetgroup.com/reports-service/reportapi"
)

func Run(ctx reportapi.Context) (map[string]any, error) {
	match := map[string]any{}

	advisorID := ctx.Query["advisor_id"]
	principalID := ctx.Query["principal_id"]

	if advisorID != "" {
		match["advisor_id"] = map[string]any{"$oid": advisorID}
	} else if principalID != "" {
		advisors, err := ctx.Mongo("broker_portal").Find("", "advisors", map[string]any{
			"$or": []any{
				map[string]any{"principal_id": principalID},
				map[string]any{"user_id": map[string]any{"$oid": principalID}},
			},
		}, map[string]any{"projection": map[string]any{"_id": 1}})
		if err != nil {
			return nil, fmt.Errorf("resolve principal's advisors: %w", err)
		}

		ids := make([]any, 0, len(advisors.Data))
		for _, a := range advisors.Data {
			ids = append(ids, map[string]any{"$oid": fmt.Sprintf("%v", a["_id"])})
		}
		if len(ids) == 0 {
			return map[string]any{"clients": map[string]any{"data": []map[string]any{}, "count": 0}}, nil
		}
		match["advisor_id"] = map[string]any{"$in": ids}
	}
	// Neither set: match stays {} -> every client (ADMIN "All").

	res, err := ctx.Mongo("broker_portal").Aggregate("", "clients", []any{
		map[string]any{"$match": match},
		map[string]any{"$lookup": map[string]any{"from": "advisors", "localField": "advisor_id", "foreignField": "_id", "as": "advisor"}},
		map[string]any{"$unwind": map[string]any{"path": "$advisor", "preserveNullAndEmptyArrays": true}},
		map[string]any{"$sort": map[string]any{"created_at": -1}},
		map[string]any{"$limit": 5000},
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{"clients": map[string]any{"data": res.Data, "count": res.Count}}, nil
}
