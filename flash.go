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
const binaryName = "sketch.elf-zsk.bin"

func flash(ctx context.Context, cmder board.Boarder, binary *paths.Path, files []*paths.Path) error {
	err := cmder.MkDirAll(ctx, binaryDir)
	if err != nil {
		return err
	}

	feedback.Logf("Pushing binary %q", binary)
	remoteBinary, err := pushBinary(ctx, cmder, binary)
	if err != nil {
		return err
	}

	feedback.Logf("Pushing config files: %v", files)
	remoteFiles, err := pushFiles(ctx, cmder, files)
	if err != nil {
		return err
	}

	args := makeOpenOCDCmd(remoteBinary, remoteFiles...)
	feedback.Logf("Running command: %s", strings.Join(args, " "))
	err = cmder.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("error running OpenOCD: %w", err)
	}

	return nil
}

func pushBinary(ctx context.Context, cmder board.Boarder, binary *paths.Path) (string, error) {
	destination := path.Join(binaryDir, binaryName)

	if err := cmder.CopyTo(ctx, binary.String(), destination); err != nil {
		return "", err
	}

	return destination, nil
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

func makeOpenOCDCmd(binary string, files ...string) []string {
	args := []string{
		openOCDBin, "-d2",
		"-s", openOCDPath,
		"-s", openOCDPath + "/share/openocd/scripts",
		"-f", "openocd_gpiod.cfg",
		"-c", "set filename " + binary,
	}
	for _, file := range files {
		args = append(args, "-f", file)
	}
	return args
}
