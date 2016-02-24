// Package babel implements the "babel" runtime.
package babel

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/apex/log"
	"github.com/jpillora/archive"

	"github.com/apex/apex/function"
	"github.com/apex/apex/plugins/nodejs"
)

// TODO(tj): ignore index.js, make an api around
// file ignoring so it's easy for plugins to alter

func init() {
	function.RegisterPlugin("babel", &Plugin{})
}

var (
	parent = &nodejs.Plugin{}
)

const (
	// Runtime name used by Apex
	Runtime = "babel"
)

// Plugin implementation.
type Plugin struct{}

// Open adds the browserify-friendly handler.
func (p *Plugin) Open(fn *function.Function) error {
	if fn.Runtime != Runtime {
		return nil
	}

	// the nodejs plugin env template uses this for its require()
	fn.Handler = "index.default"
	return nil
}

// Build executes browserify to generate main.js. This plugin "inherits"
// from the nodejs plugin in order to support the env prelude script.
func (p *Plugin) Build(fn *function.Function, zip *archive.Archive) error {
	if fn.Runtime != Runtime {
		return nil
	}

	fn.Runtime = "nodejs"

	// TODO(tj): lib-ify env portion so we can call it here instead of "inheriting"
	if err := parent.Build(fn, zip); err != nil {
		return err
	}

	inFile := "index.js"

	// .env was used, change the browserify input
	if fn.Handler == "_apex_index.handle" {
		inFile = "_apex_index.js"
	}

	// change handler to browserify file
	fn.Handler = "main.default"

	bin := "node_modules/.bin/browserify"
	in := filepath.Join(fn.Path, inFile)
	out := filepath.Join(fn.Path, "main.js")

	log.WithFields(log.Fields{
		"in":  in,
		"out": out,
	}).Debug("browserify")

	// TODO(tj): helper for nice error output, we do similar elsewhere
	cmd := exec.Command(bin, "-s", "default", "-t", "babelify", "-o", out, in)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("browserify: %s", b)
	}

	return nil
}

// Clean removes main.js.
func (p *Plugin) Clean(fn *function.Function) error {
	// TODO(tj): we could really use stateful plugins
	if fn.Runtime != "nodejs" {
		return nil
	}

	path := filepath.Join(fn.Path, "main.js")
	fn.Log.WithField("file", path).Debug("remove")
	return os.Remove(path)
}
