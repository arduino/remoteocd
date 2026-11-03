// This file is part of remoteocd.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package board

import (
	"context"
	"errors"
	"io/fs"

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

func (l *LocalCmd) CopyTo(_ context.Context, src, dst string, mode fs.FileMode) error {
	// Remove first: chmod fails on a file of another user.
	if err := paths.New(dst).Remove(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := paths.New(src).CopyTo(paths.New(dst)); err != nil {
		return err
	}
	return paths.New(dst).Chmod(mode)
}

func (l *LocalCmd) MkDirAll(_ context.Context, path string, mode fs.FileMode) error {
	p := paths.New(path)
	if err := p.MkdirAll(); err != nil {
		return err
	}
	// Skip chmod when the mode is correct: chmod fails on a folder of another user.
	if info, err := p.Stat(); err == nil && info.Mode().Perm() == mode {
		return nil
	}
	return p.Chmod(mode)
}
