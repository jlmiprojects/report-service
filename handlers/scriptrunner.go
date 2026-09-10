package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"blueassetgroup.com/reports-service/reportapi"
	rtsymbols "blueassetgroup.com/reports-service/reportapi/symbols"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// allowedStdlib is the subset of yaegi's stdlib symbol table exposed to
// report scripts. stdlib.Symbols is a blanket include (os, net, ...);
// filtering it keeps scripts inside the reportapi surface plus a handful of
// everyday packages, rather than handing them the filesystem/network.
var allowedStdlib = []string{
	"fmt/fmt",
	"strings/strings",
	"sort/sort",
	"time/time",
	"math/math",
	"strconv/strconv",
}

// ScriptRunner loads, caches and runs the .go report scripts under
// scriptDir, interpreting them via yaegi. Each script is compiled once and
// reused across requests, recompiled only when its file's mtime changes (see
// reloadable in reload.go).
type ScriptRunner struct {
	scriptDir string

	mu    sync.Mutex
	cache map[string]*reloadable[*interp.Interpreter]
}

func NewScriptRunner(scriptDir string) *ScriptRunner {
	return &ScriptRunner{scriptDir: scriptDir, cache: make(map[string]*reloadable[*interp.Interpreter])}
}

// Run loads/reuses <scriptDir>/<name>.go and calls its
// Run(reportapi.Context) (map[string]any, error) entry point.
func (sr *ScriptRunner) Run(name string, ctx reportapi.Context) (map[string]any, error) {
	i, err := sr.load(name)
	if err != nil {
		return nil, err
	}
	v, err := i.Eval("script.Run")
	if err != nil {
		return nil, fmt.Errorf("script %s.go: missing func Run(reportapi.Context) (map[string]any, error): %w", name, err)
	}
	run, ok := v.Interface().(func(reportapi.Context) (map[string]any, error))
	if !ok {
		return nil, fmt.Errorf("script %s.go: Run has the wrong signature (want func(reportapi.Context) (map[string]any, error))", name)
	}
	return run(ctx)
}

// RunOptions loads/reuses <scriptDir>/<name>.go and calls its
// Options(reportapi.Context) ([]reportapi.Option, error) entry point, used
// for "lookup" report parameters.
func (sr *ScriptRunner) RunOptions(name string, ctx reportapi.Context) ([]reportapi.Option, error) {
	i, err := sr.load(name)
	if err != nil {
		return nil, err
	}
	v, err := i.Eval("script.Options")
	if err != nil {
		return nil, fmt.Errorf("script %s.go: missing func Options(reportapi.Context) ([]reportapi.Option, error): %w", name, err)
	}
	options, ok := v.Interface().(func(reportapi.Context) ([]reportapi.Option, error))
	if !ok {
		return nil, fmt.Errorf("script %s.go: Options has the wrong signature (want func(reportapi.Context) ([]reportapi.Option, error))", name)
	}
	return options(ctx)
}

// load returns a ready-to-call interpreter for <scriptDir>/<name>.go,
// (re)compiling it if the file has changed since the last call.
func (sr *ScriptRunner) load(name string) (*interp.Interpreter, error) {
	path := filepath.Join(sr.scriptDir, name+".go")

	sr.mu.Lock()
	entry, ok := sr.cache[name]
	if !ok {
		entry = newReloadable(
			func() (any, error) { return statFileSig(path) },
			func() (*interp.Interpreter, error) { return compileScript(path) },
		)
		sr.cache[name] = entry
	}
	sr.mu.Unlock()

	return entry.Get()
}

func compileScript(path string) (*interp.Interpreter, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	i := interp.New(interp.Options{})
	if err := i.Use(filteredStdlib()); err != nil {
		return nil, fmt.Errorf("script %s: failed to load stdlib symbols: %w", path, err)
	}
	if err := i.Use(rtsymbols.Symbols); err != nil {
		return nil, fmt.Errorf("script %s: failed to load reportapi symbols: %w", path, err)
	}

	if _, err := i.Eval(string(src)); err != nil {
		return nil, fmt.Errorf("compile script %s: %w", path, err)
	}
	return i, nil
}

// filteredStdlib returns the allow-listed subset of stdlib.Symbols (see
// allowedStdlib).
func filteredStdlib() interp.Exports {
	out := make(interp.Exports, len(allowedStdlib))
	for _, key := range allowedStdlib {
		if syms, ok := stdlib.Symbols[key]; ok {
			out[key] = syms
		}
	}
	return out
}
