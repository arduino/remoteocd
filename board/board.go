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
	"bytes"
	"context"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

type Boarder interface {
	Run(ctx context.Context, args ...string) error
	CopyTo(ctx context.Context, src, dst string) error
	MkDirAll(ctx context.Context, path string) error
}

var knownBoards = []string{"arduino,imola", "arduino,monza", "arduino"}

var OnBoard = sync.OnceValue(func() bool {
	trimAndLower := func(s []byte) []byte {
		return bytes.ToLower(bytes.Trim(s, " \n\t\r\x00"))
	}

	readFile := func(path string) ([]byte, error) {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return io.ReadAll(f)
	}

	// legacy check for imola
	if buf, err := readFile("/sys/class/dmi/id/product_name"); err == nil {
		return string(trimAndLower(buf)) == "imola"
	}

	var compatibles []string
	if buf, err := readFile("/sys/firmware/devicetree/base/compatible"); err == nil {
if buf, err := readFile("/sys/firmware/devicetree/base/compatible"); err == nil {
	for _, raw := range bytes.Split(buf, []byte{'\x00'}) {
		compatible := string(trimAndLower(raw))

		for _, knownBoard := range knownBoards {
			if strings.HasPrefix(compatible, knownBoard) {
				return true
			}
		}
	}
}

	for _, knownBoard := range knownBoards {
		for _, compatible := range compatibles {
			if strings.HasPrefix(compatible, knownBoard) {
				return true
			}
		}
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
