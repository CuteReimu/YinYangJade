package scripts

import (
	"embed"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
)

//go:embed *.py
var scriptFS embed.FS

//go:embed requirements.txt
var requirements []byte

var e error

func RunPythonScript(scriptName string, args ...string) ([]byte, error) {
	const scriptDir = "scripts_hkbot"
	depOnce.Do(func() {
		files, err := scriptFS.ReadDir(".")
		if err != nil {
			e = errors.WithStack(err)
			return
		}

		_ = os.RemoveAll(scriptDir)
		_ = os.MkdirAll(scriptDir, 0700)

		for _, name := range files {
			data, err := scriptFS.ReadFile(name.Name())
			if err != nil {
				e = errors.WithStack(err)
				return
			}
			err = os.WriteFile(filepath.Join(scriptDir, name.Name()), data, 0600)
			if err != nil {
				e = errors.WithStack(err)
				return
			}
		}

		// 提取依赖库到临时目录
		err = extractDependencies("libs")
		if err != nil {
			e = errors.WithStack(err)
		}
	})
	if e != nil {
		return nil, e
	}

	// 执行Python脚本
	cmd := exec.Command("python3", append([]string{filepath.Join(scriptDir, scriptName)}, args...)...)
	cmd.Env = append(cmd.Environ(), "PYTHONPATH=libs")
	slog.Debug("Executing Python script", "cmd", cmd)
	return cmd.Output()
}

var depOnce sync.Once

func extractDependencies(destDir string) error {
	err := os.WriteFile("requirements_hkbot.txt", requirements, 0600)
	if err != nil {
		slog.Error("Unable to write requirements_hkbot.txt", "err", err)
		return errors.WithStack(err)
	}
	// 执行Python脚本
	cmd := exec.Command("pip3", "install", "-t", destDir, "-r", "requirements_hkbot.txt")
	slog.Debug("Executing Python script", "cmd", cmd)
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Error executing Python script", "output", string(output), "err", err)
		return errors.WithStack(err)
	}
	return nil
}
