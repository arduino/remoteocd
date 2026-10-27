// This file is part of remoteocd.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/remoteocd/board"
	"github.com/arduino/remoteocd/feedback"
)

func flash(ctx context.Context, cmder board.Boarder, binaries paths.PathList, configs paths.PathList) error {
	binaryDir, err := cmder.MkTempDir(ctx)
	if err != nil {
		return err
	}
	defer func() {
		// Without cancel: cleanup must run also after an interrupt.
		if err := cmder.RemoveAll(context.WithoutCancel(ctx), binaryDir); err != nil {
			feedback.Logf("Failed to remove %s: %v", binaryDir, err)
		}
	}()

	feedback.Logf("Pushing binary: %v", binaries)
	remoteBinaries, err := pushFiles(ctx, cmder, binaryDir, binaries)
	if err != nil {
		return err
	}

	feedback.Logf("Pushing config files: %v", configs)
	remoteConfigs, err := pushFiles(ctx, cmder, binaryDir, configs)
	if err != nil {
		return err
	}

	args := makeOpenOCDCmd(remoteBinaries, remoteConfigs)
	feedback.Logf("Running command: %s", strings.Join(args, " "))
	err = cmder.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("error running OpenOCD: %w", err)
	}

	return nil
}

func pushFiles(ctx context.Context, cmder board.Boarder, binaryDir string, files []*paths.Path) ([]string, error) {
	remoteFiles := make([]string, 0, len(files))
	for _, file := range files {
		destination := path.Join(binaryDir, file.Base())

		if err := cmder.CopyTo(ctx, file.String(), destination); err != nil {
			return nil, err
		}

		remoteFiles = append(remoteFiles, destination)
	}

	return remoteFiles, nil
}

const openOCDPath = "/opt/openocd"
const openOCDBin = openOCDPath + "/bin/openocd"

func makeOpenOCDCmd(binaries []string, configs []string) []string {
	args := []string{
		openOCDBin, "-d2",
		"-s", openOCDPath,
		"-s", openOCDPath + "/share/openocd/scripts",
		"-f", "openocd_gpiod.cfg",
		"-c", "set filename " + binaries[0], // backward compatibility: first binary is set also as "filename"
	}
	for i, binary := range binaries {
		args = append(args, "-c", fmt.Sprintf("set filename%d %s", i, binary))
	}
	for _, file := range configs {
		args = append(args, "-f", file)
	}
	return args
}
