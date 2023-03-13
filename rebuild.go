package gone

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// DEFAULT_BUILD_NAME 编译后的文件名
const DEFAULT_BUILD_NAME = "main"

// gbuild 编译主文件
func gbuild(basePath string, mainFile string) error {
	cmd := exec.Command("go", "build", "-o", filepath.Join(basePath, ".gone", DEFAULT_BUILD_NAME), mainFile)
	if err := cmd.Start(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

// grun 执行编译后的文件
func grun(basePath string) (*exec.Cmd, error) {
	cmd := exec.Command(filepath.Join(basePath, ".gone", DEFAULT_BUILD_NAME))
	cmd.Env = append(cmd.Env, "GONE_ROUTINE=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	return cmd, nil
}
