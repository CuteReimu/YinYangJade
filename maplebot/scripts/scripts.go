package scripts

import (
	"embed"
	"log/slog"
	"os"
	"os/exec"
	"sync"
)

//go:embed *.py
var scriptFS embed.FS

//go:embed requirements.txt
var requirements []byte

func RunPythonScript(scriptName string, args ...string) ([]byte, error) {
	// 读取嵌入的Python文件
	data, err := scriptFS.ReadFile(scriptName)
	if err != nil {
		return nil, err
	}

	// 创建临时文件
	tmpfile, err := os.CreateTemp("", "script-*.py")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	// 写入Python代码
	closed := false
	defer func() {
		if !closed {
			_ = tmpfile.Close()
		}
	}()
	if _, err := tmpfile.Write(data); err != nil {
		return nil, err
	}
	if err := tmpfile.Close(); err != nil {
		return nil, err
	}
	closed = true

	// 提取依赖库到临时目录
	extractDependencies("libs")

	// 执行Python脚本
	cmd := exec.Command("python3", append([]string{scriptName}, args...)...)
	cmd.Env = append(cmd.Environ(), "PYTHONPATH=libs")
	slog.Debug("Executing Python script", "cmd", cmd)
	return cmd.Output()
}

var depOnce sync.Once

func extractDependencies(destDir string) {
	depOnce.Do(func() {
		err := os.WriteFile("requirements.txt", requirements, 0600)
		if err != nil {
			slog.Error("Unable to write requirements.txt", "err", err)
			return
		}
		// 执行Python脚本
		cmd := exec.Command("pip3", "install", "-t", destDir, "-r", "requirements.txt")
		slog.Debug("Executing Python script", "cmd", cmd)
		output, err := cmd.Output()
		if err != nil {
			slog.Error("Error executing Python script", "output", string(output), "err", err)
		}
	})
}
