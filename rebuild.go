package gone

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// build 编译文件
func build(basePath string) bool {
	cmd := exec.Command("go", "build", "-o", filepath.Join(basePath, ".gone", "main"), "main.go")
	if err := cmd.Start(); err != nil {
		return false
	}
	if err := cmd.Wait(); err != nil {
		return false
	}
	return true
}

func run(basePath string) *exec.Cmd {
	cmd := exec.Command(filepath.Join(basePath, ".gone", "main"))
	cmd.Env = append(cmd.Env, "GONE_RUNTIME=1")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	return cmd
}
