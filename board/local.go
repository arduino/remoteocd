// This file is part of remoteocd.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

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
