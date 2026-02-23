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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/arduino/go-paths-helper"

	"github.com/arduino/remoteocd/board"
	"github.com/arduino/remoteocd/feedback"
)

const binaryDir = "/tmp/remoteocd/"
const binaryName = "sketch.elf-zsk.bin"

const stateDir = "/var/tmp/remoteocd/"
const stateFile = "last_flash_state.sha256"

func flash(ctx context.Context, cmder board.Boarder, binary *paths.Path, files []*paths.Path) error {
	shaState, err := computeSha256(binary, files)
	if err != nil {
		return fmt.Errorf("failed to compute flash state hash: %w", err)
	}

	lastHash, err := pullLastFlashStateHash(ctx, cmder)
	if err == nil && lastHash == shaState {
		feedback.Printf("Skipping upload: binary and config files unchanged (hash: %s)", shaState)
		return nil
	}

	err = cmder.MkDirAll(ctx, binaryDir)
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

	// Save new state hash to board
	destHashFile, err := pushFlashStateHash(ctx, cmder, shaState)
	if err != nil {
		feedback.Printf("warning: failed to push flash state hash: %v", err)
	}
	feedback.Printf("State hash file saved correctly into %v", destHashFile)

	return nil
}

// computeSha256 computes a hash of the binary and config file names and contents
func computeSha256(binary *paths.Path, files []*paths.Path) (string, error) {
	h := sha256.New()

	// Binary name
	binaryName := binary.Base()
	h.Write([]byte(binaryName))

	// Binary content
	f, err := binary.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	// Config files: name and content
	for _, file := range files {
		fileBaseName := file.Base()
		h.Write([]byte(fileBaseName))
		f, err := file.Open()
		if err != nil {
			return "", err
		}
		defer f.Close()
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func pullLastFlashStateHash(ctx context.Context, cmder board.Boarder) (string, error) {
	localTmp := path.Join(os.TempDir(), stateFile)
	remote := path.Join(stateDir, stateFile)
	if err := cmder.PullFrom(ctx, remote, localTmp); err != nil {
		return "", err
	}
	data, err := os.ReadFile(localTmp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func pushFlashStateHash(ctx context.Context, cmder board.Boarder, hash string) (string, error) {
	localTmp := paths.New(os.TempDir(), stateFile)
	if err := localTmp.WriteFile([]byte(hash)); err != nil {
		return "", err
	}

	dest := path.Join(stateDir, stateFile)
	if err := cmder.CopyTo(ctx, localTmp.String(), dest); err != nil {
		return "", err
	}

	return dest, nil
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
