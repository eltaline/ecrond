package main

import (
	"os/exec"
)

// QuickExec quick exec an simple command line
func QuickExec(cmdLine string, workDir ...string) (string, error) {
	return ExecLine(cmdLine, workDir...)
}

// ExecLine quick exec an command line string
func ExecLine(cmdLine string, workDir ...string) (string, error) {
	p := NewParser(cmdLine)

	// create a new Cmd instance
	cmd := p.NewExecCmd()
	if len(workDir) > 0 {
		cmd.Dir = workDir[0]
	}

	bs, err := cmd.CombinedOutput()
	return string(bs), err
}

// ExecCmd an command and return output
func ExecCmd(binName string, args []string, workDir ...string) (string, error) {
	cmd := exec.Command(binName, args...)
	if len(workDir) > 0 {
		cmd.Dir = workDir[0]
	}

	bs, err := cmd.CombinedOutput()
	return string(bs), err
}

// ShellExec exec command by shell
func ShellExec(cmdLine string, shells ...string) (string, error) {
	shell := "/bin/bash"
	if len(shells) > 0 {
		shell = shells[0]
	}

	out, err := exec.Command(shell, "-c", cmdLine).CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}
