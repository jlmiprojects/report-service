// Package symbols is the yaegi symbol table exposing reportapi's exported API
// to interpreted report scripts (see handlers/scriptrunner.go). The
// accompanying generated file is produced by `go generate ./reportapi/...`
// and must be re-run after any change to reportapi's public API.
//
//go:generate go run github.com/traefik/yaegi/cmd/yaegi extract -name symbols blueassetgroup.com/reports-service/reportapi
package symbols

import "reflect"

// Symbols is populated by the generated file in this package — DO NOT EDIT
// that file directly.
var Symbols = map[string]map[string]reflect.Value{}
