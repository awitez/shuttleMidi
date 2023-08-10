package notification

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

type commander struct {
	path string
	c    string
}

func newCommander() *commander {
	return &commander{path: os.Getenv("SHELL"), c: "-c"}
}

func (cmdr *commander) exec(osascript string) (output string, err error) {
	args := append([]string{cmdr.c}, osascript)
	cmd := exec.Command(cmdr.path, args...)

	cmd.SysProcAttr = &syscall.SysProcAttr{}

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer func() { _ = stdoutPipe.Close() }()

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	defer func() { _ = stderrPipe.Close() }()

	if err = cmd.Start(); err != nil {
		return "", err
	}

	outputBuffer := new(bytes.Buffer)
	stderr, err := ioutil.ReadAll(stderrPipe)
	if err != nil {
		return "", err
	}
	if len(stderr) != 0 {
		outputBuffer.Write(stderr)
	}

	stdout, err := ioutil.ReadAll(stdoutPipe)
	if err != nil {
		return outputBuffer.String(), err
	}
	if len(stdout) != 0 {
		outputBuffer.Write(stdout)
	}

	_ = cmd.Wait()
	return outputBuffer.String(), nil
}
