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

const binaryDir = "/tmp/remoteocd/"

func flash(ctx context.Context, cmder board.Boarder, binaries paths.PathList, configs paths.PathList) error {
	err := cmder.MkDirAll(ctx, binaryDir)
	if err != nil {
		return err
	}

	feedback.Logf("Pushing binary: %v", binaries)
	remoteBinaries, err := pushFiles(ctx, cmder, binaries)
	if err != nil {
		return err
	}

	feedback.Logf("Pushing config files: %v", configs)
	remoteConfigs, err := pushFiles(ctx, cmder, configs)
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

func pushFiles(ctx context.Context, cmder board.Boarder, files []*paths.Path) ([]string, error) {
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
