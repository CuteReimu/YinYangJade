package scripts

import (
	"embed"

	"github.com/CuteReimu/YinYangJade/scriptutil"
)

//go:embed *.py
var scriptFS embed.FS

//go:embed requirements.txt
var requirements []byte

var runner = scriptutil.NewPythonScriptRunner(scriptFS, requirements, "scripts", "requirements.txt")

func RunPythonScript(scriptName string, args ...string) ([]byte, error) {
	return runner.RunScript(scriptName, args...)
}
