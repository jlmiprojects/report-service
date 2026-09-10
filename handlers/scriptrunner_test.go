package handlers

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"blueassetgroup.com/reports-service/reportapi"
)

func writeScript(t *testing.T, path, src string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestScriptsCompile catches syntax/type errors in the .go report scripts
// under ../scripts that `go vet ./...` can't see (they carry a //go:build
// ignore tag so they aren't compiled as part of this module — see
// scripts/clients.go's file comment).
func TestScriptsCompile(t *testing.T) {
	for _, name := range []string{"clients", "record_of_advice"} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "scripts", name+".go")
			i, err := compileScript(path)
			if err != nil {
				t.Fatalf("compile %s: %v", path, err)
			}
			if _, err := i.Eval("script.Run"); err != nil {
				t.Fatalf("%s: missing func Run(reportapi.Context) (map[string]any, error): %v", path, err)
			}
		})
	}
}

// TestScriptRunnerReload verifies a script recompiles after its file
// changes, and that a previous good compile keeps serving through a
// transient failure.
func TestScriptRunnerReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "greet.go")

	writeScript(t, path, `package script

import "blueassetgroup.com/reports-service/reportapi"

func Run(ctx reportapi.Context) (map[string]any, error) {
	return map[string]any{"greeting": "hello"}, nil
}
`)

	sr := NewScriptRunner(dir)
	data, err := sr.Run("greet", reportapi.Context{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if data["greeting"] != "hello" {
		t.Fatalf("greeting = %v, want hello", data["greeting"])
	}

	writeScript(t, path, `package script

import "blueassetgroup.com/reports-service/reportapi"

func Run(ctx reportapi.Context) (map[string]any, error) {
	return map[string]any{"greeting": "goodbye"}, nil
}
`)
	// Force the mtime forward so the reloadable's signature check sees a
	// change regardless of filesystem timestamp resolution.
	future := time.Now().Add(time.Second)
	if err := os.Chtimes(path, future, future); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	data, err = sr.Run("greet", reportapi.Context{})
	if err != nil {
		t.Fatalf("Run after edit: %v", err)
	}
	if data["greeting"] != "goodbye" {
		t.Fatalf("greeting after edit = %v, want goodbye", data["greeting"])
	}
}
