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

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/remoteocd/feedback"
)

var _ Boarder = (*LocalCmd)(nil)

type LocalCmd struct{}

func (l *LocalCmd) Run(ctx context.Context, args ...string) error {
	cmd, err := paths.NewProcess(nil, args...)
	if err != nil {
		return err
	}

	cmd.RedirectStderrTo(feedback.GetStdout())
	cmd.RedirectStdoutTo(feedback.GetStdout())

	return cmd.RunWithinContext(ctx)
}

func (l *LocalCmd) CopyTo(_ context.Context, src, dst string) error {
	return paths.New(src).CopyTo(paths.New(dst))
}

func (l *LocalCmd) MkDirAll(_ context.Context, path string) error {
	return paths.New(path).MkdirAll()
}

// PullFrom copies a file locally
func (l *LocalCmd) PullFrom(_ context.Context, src, dst string) error {
	return paths.New(src).CopyTo(paths.New(dst))
}
