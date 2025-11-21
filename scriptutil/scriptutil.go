package scriptutil

import (
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
)

// PythonScriptRunner 管理 Python 脚本的执行环境
type PythonScriptRunner struct {
	scriptFS         fs.FS
	requirements     []byte
	scriptDir        string
	requirementsFile string
	libsDir          string
	initOnce         sync.Once
	initErr          error
}

// NewPythonScriptRunner 创建一个新的 Python 脚本运行器
func NewPythonScriptRunner(scriptFS fs.FS, requirements []byte, scriptDir, requirementsFile string) *PythonScriptRunner {
	return &PythonScriptRunner{
		scriptFS:         scriptFS,
		requirements:     requirements,
		scriptDir:        scriptDir,
		requirementsFile: requirementsFile,
		libsDir:          "libs",
	}
}

// RunScript 执行指定的 Python 脚本
func (r *PythonScriptRunner) RunScript(scriptName string, args ...string) ([]byte, error) {
	r.initOnce.Do(func() {
		r.initErr = r.initialize()
	})
	if r.initErr != nil {
		return nil, r.initErr
	}

	// 执行Python脚本
	cmd := exec.Command("python3", append([]string{filepath.Join(r.scriptDir, scriptName)}, args...)...)
	cmd.Env = append(cmd.Environ(), "PYTHONPATH="+r.libsDir)
	slog.Debug("Executing Python script", "cmd", cmd)
	return cmd.CombinedOutput()
}

// initialize 初始化脚本执行环境
func (r *PythonScriptRunner) initialize() error {
	files, err := fs.ReadDir(r.scriptFS, ".")
	if err != nil {
		return errors.WithStack(err)
	}

	_ = os.RemoveAll(r.scriptDir)
	if err := os.MkdirAll(r.scriptDir, 0700); err != nil {
		return errors.WithStack(err)
	}

	for _, name := range files {
		data, err := fs.ReadFile(r.scriptFS, name.Name())
		if err != nil {
			return errors.WithStack(err)
		}
		err = os.WriteFile(filepath.Join(r.scriptDir, name.Name()), data, 0600)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	// 提取依赖库到临时目录
	if err := r.extractDependencies(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// extractDependencies 提取 Python 依赖到指定目录
func (r *PythonScriptRunner) extractDependencies() error {
	err := os.WriteFile(r.requirementsFile, r.requirements, 0600)
	if err != nil {
		slog.Error("Unable to write requirements file", "file", r.requirementsFile, "err", err)
		return errors.WithStack(err)
	}

	cmd := exec.Command("pip3", "install", "-t", r.libsDir, "-r", r.requirementsFile)
	slog.Debug("Installing Python dependencies", "cmd", cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("Error installing Python dependencies", "output", string(output), "err", err)
		return errors.WithStack(err)
	}
	return nil
}
