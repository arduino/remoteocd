// This file is part of arduino-flasher-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

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
