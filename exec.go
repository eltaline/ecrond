/*
 * Copyright © 2022 Andrey Kuvshinov. Contacts: <syslinux@protonmail.com>
 * Copyright © 2022 Eltaline OU. Contacts: <eltaline.ou@gmail.com>
 * Copyright © 2016 inhere
 *
 * This file is part of eCrond.
 *
 * eCrond is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * eCrond is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

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
