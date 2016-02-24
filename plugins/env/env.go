// Package env populates .env.json if the function has any environment variables defined.
package env

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/jpillora/archive"

	"github.com/apex/apex/function"
)

func init() {
	function.RegisterPlugin("env", &Plugin{})
}

// Filename of file with environment variables.
const Filename = ".env.json"

// Plugin implementation.
type Plugin struct{}

// Build hook adds .env.json populate with Function.Enironment.
func (p *Plugin) Build(fn *function.Function, zip *archive.Archive) error {
	if len(fn.Environment) == 0 {
		return nil
	}

	fn.Log.WithField("env", fn.Environment).Debug("adding env")

	b, err := json.Marshal(fn.Environment)
	if err != nil {
		return err
	}

	path := filepath.Join(fn.Path, Filename)
	fn.Log.WithField("file", path).Debug("create")
	return ioutil.WriteFile(path, b, 0666)
}

// Clean removes the environment file.
func (p *Plugin) Clean(fn *function.Function) error {
	if len(fn.Environment) == 0 {
		return nil
	}

	path := filepath.Join(fn.Path, Filename)
	fn.Log.WithField("file", path).Debug("remove")
	return os.Remove(path)
}
