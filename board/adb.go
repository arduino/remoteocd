// This file is part of remoteocd.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package board

import (
	"context"
	"fmt"
	"io/fs"

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

func (a *ADBCmd) CopyTo(ctx context.Context, src, dst string, mode fs.FileMode) error {
	p, err := paths.NewProcess(nil, a.ADBPath, "-s", a.Serial, "push", src, dst)
	if err != nil {
		return err
	}
	out, err := p.RunAndCaptureCombinedOutput(ctx)
	if err != nil {
		return fmt.Errorf("copy files error: %w: %s", err, out)
	}

	p, err = paths.NewProcess(nil, a.ADBPath, "-s", a.Serial, "shell", "chmod", fmt.Sprintf("%o", mode), shellQuote(dst))
	if err != nil {
		return err
	}
	out, err = p.RunAndCaptureCombinedOutput(ctx)
	if err != nil {
		return fmt.Errorf("chmod error: %w: %s", err, out)
	}
	return nil
}

func (a *ADBCmd) MkDirAll(ctx context.Context, path string, mode fs.FileMode) error {
	p, err := paths.NewProcess(nil, a.ADBPath, "-s", a.Serial, "shell", "install", "-d", "-m", fmt.Sprintf("%o", mode), shellQuote(path))
	if err != nil {
		return err
	}
	out, err := p.RunAndCaptureCombinedOutput(ctx)
	if err != nil {
		return fmt.Errorf("makedir error: %w: %s", err, out)
	}
	return nil
}
