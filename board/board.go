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

package board

import (
	"context"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
)

type Boarder interface {
	Run(ctx context.Context, args ...string) error
	CopyTo(ctx context.Context, src, dst string) error
	MkDirAll(ctx context.Context, path string) error
}

var OnBoard = sync.OnceValue(func() bool {
	var boardNames = []string{"UNO Q\n", "Imola\n", "Inc. Robotics RB1\n"}
	buf, err := os.ReadFile("/sys/class/dmi/id/product_name")
	if err == nil && slices.Contains(boardNames, string(buf)) {
		return true
	}
	return false
})()

// escapeArgs escapes arguments that contain spaces by wrapping them in quotes.
// This should allow argument forwarding on a remote shells.
func escapeArgs(args []string) []string {
	for i, arg := range args {
		if strings.Contains(arg, " ") {
			args[i] = strconv.Quote(arg)
		}
	}
	return args
}
