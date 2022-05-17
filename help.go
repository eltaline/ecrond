/*

Copyright © 2020 Andrey Kuvshinov. Contacts: <syslinux@protonmail.com>
Copyright © 2020 Eltaline OU. Contacts: <eltaline.ou@gmail.com>
All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

The ecrond project contains unmodified/modified libraries imports too with
separate copyright notices and license terms. Your use of the source code
this libraries is subject to the terms and conditions of licenses these libraries.

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
