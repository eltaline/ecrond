/*
 * Copyright © 2022 Andrey Kuvshinov. Contacts: <syslinux@protonmail.com>
 * Copyright © 2022 Eltaline OU. Contacts: <eltaline.ou@gmail.com>
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
	"fmt"
	"os"
)

// FileExists : Check existence of requested file
func FileExists(filename string) bool {

	if finfo, err := os.Stat(filename); err == nil {

		if finfo.Mode().IsRegular() {
			return true
		}

	}

	return false

}

// DirExists : Check existence of requested directory
func DirExists(filename string) bool {

	if fi, err := os.Stat(filename); err == nil {

		if fi.Mode().IsDir() {
			return true
		}

	}

	return false

}

// GetPID : Get current pid number and return int and string representation of pid
func GetPID() (gpid string, fpid string) {

	gpid = fmt.Sprintf("%d", os.Getpid())
	fpid = fmt.Sprintf("%s\n", gpid)

	return gpid, fpid

}
