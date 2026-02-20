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
	// Compute current state hash
	stateHash, err := computeFlashStateHash(binary, files)
	if err != nil {
		return fmt.Errorf("failed to compute flash state hash: %w", err)
	}

	// Pull last state hash from board
	lastHash, err := pullLastFlashStateHash(ctx, cmder)
	if err == nil && lastHash == stateHash {
		feedback.Logf("Skipping flash: binary and config files unchanged (hash: %s)", stateHash)
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
	if err = pushFlashStateHash(ctx, cmder, stateHash); err != nil {
		feedback.Printf("warning: failed to push flash state hash: %v", err)
	}

	return nil
}

// computeFlashStateHash computes a hash of the binary and config file names and contents
func computeFlashStateHash(binary *paths.Path, files []*paths.Path) (string, error) {
	h := sha256.New()
	// Binary name
	h.Write([]byte(binary.Base()))
	// Binary content
	f, err := binary.Open()
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(h, f); err != nil {
		f.Close()
		return "", err
	}
	f.Close()
	// Config files: name and content
	for _, file := range files {
		h.Write([]byte(file.Base()))
		f, err := file.Open()
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(h, f); err != nil {
			f.Close()
			return "", err
		}
		f.Close()
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// pullLastFlashStateHash tries to pull the last flash state hash from the board
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

// pushFlashStateHash saves the new state hash to the board
func pushFlashStateHash(ctx context.Context, cmder board.Boarder, hash string) error {
	localTmp := path.Join(os.TempDir(), stateFile)
	if err := os.WriteFile(localTmp, []byte(hash), 0644); err != nil {
		return err
	}
	remote := path.Join(stateDir, stateFile)
	if err := cmder.MkDirAll(ctx, stateDir); err != nil {
		return err
	}
	return cmder.CopyTo(ctx, localTmp, remote)
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

func pushHash(ctx context.Context, cmder board.Boarder, binary *paths.Path) error {
	fmt.Printf("Calculating hash for binary %q", binary)
	binaryName := binary.Base()
	f, err := binary.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	binaryHashName := binaryName + ".sha256"
	tmp, err := paths.MkTempFile(nil, binaryHashName)
	if err != nil {
		return err
	}
	defer tmp.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, f)
	if err != nil {
		return err
	}
	_ = f.Close()
	_, err = tmp.Write(hash.Sum(nil))
	if err != nil {
		return err
	}
	_ = tmp.Close()

	destination := path.Join(binaryHashDir, binaryHashName)
	if err := cmder.MkDirAll(ctx, binaryHashDir); err != nil {
		return err
	}
	return cmder.CopyTo(ctx, tmp.Name(), destination)
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
