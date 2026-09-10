package handlers

import (
	"errors"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
)

// reloadable lazily (re)builds a value of type T, recompiling only when a
// cheap signature probe (e.g. a file's mtime) says the on-disk source has
// changed since the last build. Get is safe for concurrent use: the common
// case (no change) only runs the cheap signature probe and reads an atomic
// pointer, never the expensive build; a detected change rebuilds under a
// mutex that serializes rebuilds, not readers.
type reloadable[T any] struct {
	// signature is a cheap, frequently-run probe of the on-disk source (e.g.
	// os.Stat one file, or glob+stat a directory) returning a comparable
	// value. It must NOT do the expensive work build does.
	signature func() (any, error)
	// build does the expensive (re)compile/(re)parse, run only when
	// signature's result differs from what's cached.
	build func() (T, error)

	mu      sync.Mutex
	sig     atomic.Value
	current atomic.Pointer[T]
}

func newReloadable[T any](signature func() (any, error), build func() (T, error)) *reloadable[T] {
	return &reloadable[T]{signature: signature, build: build}
}

// Get returns the current value, rebuilding first if the signature changed.
// If a rebuild fails and a previous good value exists, that value is
// returned alongside the error so callers can degrade gracefully (serve the
// last-good template/script) instead of crashing on a transient/bad edit.
func (r *reloadable[T]) Get() (T, error) {
	sig, sigErr := r.signature()
	if sigErr == nil && r.unchanged(sig) {
		if p := r.current.Load(); p != nil {
			return *p, nil
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Re-check: another goroutine may have already rebuilt while we waited
	// for the lock.
	if sigErr == nil && r.unchanged(sig) {
		if p := r.current.Load(); p != nil {
			return *p, nil
		}
	}

	value, err := r.build()
	if err != nil {
		if p := r.current.Load(); p != nil {
			return *p, err
		}
		var zero T
		return zero, err
	}

	r.current.Store(&value)
	if sigErr == nil {
		r.sig.Store(sig)
	}
	return value, nil
}

func (r *reloadable[T]) unchanged(sig any) bool {
	cached := r.sig.Load()
	return cached != nil && reflect.DeepEqual(cached, sig)
}

// fileSig is a comparable (path, mtime) pair used as (part of) a signature.
type fileSig struct {
	Path  string
	Mtime int64
}

// statFileSig stats one file and returns its signature.
func statFileSig(path string) (fileSig, error) {
	info, err := os.Stat(path)
	if err != nil {
		return fileSig{}, err
	}
	return fileSig{Path: path, Mtime: info.ModTime().UnixNano()}, nil
}

// globSig stats every file matching pattern and returns a sorted signature
// covering all of them, so an added/removed/changed file is always detected
// (a plain "max mtime" would miss a removal).
func globSig(pattern string) ([]fileSig, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, errors.New("reload: no files match " + pattern)
	}
	sort.Strings(matches)
	sigs := make([]fileSig, len(matches))
	for i, m := range matches {
		s, err := statFileSig(m)
		if err != nil {
			return nil, err
		}
		sigs[i] = s
	}
	return sigs, nil
}

// newTemplateReloadable builds a reloadable *template.Template tree from
// every *.html file under dir (glob-and-parse, same as the old
// template.Must(...ParseGlob(...)) at startup), rebuilding whenever any
// matched file's mtime changes — including a file that's only {{define}}d
// and included by others (e.g. a shared styles partial), since the signature
// covers every matched file, not just the ones named directly.
func newTemplateReloadable(dir string, funcMap template.FuncMap) *reloadable[*template.Template] {
	pattern := dir + "/*.html"
	return newReloadable(
		func() (any, error) { return globSig(pattern) },
		func() (*template.Template, error) {
			return template.New("template").Funcs(funcMap).ParseGlob(pattern)
		},
	)
}
