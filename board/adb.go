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
	"fmt"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/remoteocd/feedback"
)

var _ Boarder = (*ADBCmd)(nil)

type ADBCmd struct {
	Serial  string
	ADBPath string
}

func (a *ADBCmd) Run(ctx context.Context, args ...string) error {
	args = escapeArgs(args)

	adbArgs := make([]string, 0, len(args)+4)
	adbArgs = append(adbArgs, a.ADBPath, "-s", a.Serial, "shell")
	adbArgs = append(adbArgs, args...)

	cmd, err := paths.NewProcess(nil, adbArgs...)
	if err != nil {
		return err
	}

	cmd.RedirectStderrTo(feedback.GetStdout())
	cmd.RedirectStdoutTo(feedback.GetStdout())

	return cmd.RunWithinContext(ctx)
}

func (a *ADBCmd) CopyTo(ctx context.Context, src, dst string) error {
	p, err := paths.NewProcess(nil, a.ADBPath, "-s", a.Serial, "push", src, dst)
	if err != nil {
		return err
	}
	out, err := p.RunAndCaptureCombinedOutput(ctx)
	if err != nil {
		return fmt.Errorf("copy files error: %w: %s", err, out)
	}
	return nil
}

func (a *ADBCmd) MkDirAll(ctx context.Context, path string) error {
	p, err := paths.NewProcess(nil, a.ADBPath, "-s", a.Serial, "shell", "mkdir", "-p", path)
	if err != nil {
		return err
	}
	out, err := p.RunAndCaptureCombinedOutput(ctx)
	if err != nil {
		return fmt.Errorf("makedir error: %w: %s", err, out)
	}
	return nil
}

// PullFrom copies a file from the board to the local system using adb pull
func (a *ADBCmd) PullFrom(ctx context.Context, src, dst string) error {
	p, err := paths.NewProcess(nil, a.ADBPath, "-s", a.Serial, "pull", src, dst)
	if err != nil {
		return err
	}
	out, err := p.RunAndCaptureCombinedOutput(ctx)
	if err != nil {
		return fmt.Errorf("copy from error: %w: %s", err, out)
	}
	return nil
}
