// This file is part of remoteocd.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of remoteocd.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package feedback

import (
	"fmt"
	"io"
	"os"
)

var verbose bool
var quiet bool

func SetVerbose(v bool) {
	verbose = v
}

func SetQuiet(q bool) {
	quiet = q
}

func Printf(format string, a ...any) {
	if !quiet {
		fmt.Printf(format+"\n", a...)
	}
}

func Logf(format string, a ...any) {
	if verbose && !quiet {
		fmt.Printf(format+"\n", a...)
	}
}

func GetStdout() io.Writer {
	if quiet {
		return io.Discard
	}
	return os.Stdout
}
