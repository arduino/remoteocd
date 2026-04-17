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
	"os"

	"github.com/arduino/go-paths-helper"
	"github.com/spf13/cobra"
	"go.bug.st/cleanup"

	"github.com/arduino/remoteocd/board"
	"github.com/arduino/remoteocd/feedback"
)

// Version is set at build time using -ldflags "-X 'main.Version=1.0.0'"
var Version = "0.0.0-dev"

func main() {
	rootCmd := newRootCmd()
	ctx := context.Background()
	ctx, _ = cleanup.InterruptableContext(ctx)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var verbose bool
	var quite bool
	root := &cobra.Command{
		Use:   "remoteocd",
		Short: "A CLI tool to upload a firmaware to the Uno Q microcontroller",
		Long: `This tool is able to upload a firmware on the microcontroller either from the board itself or from an host PC.
It uses OpenOCD to flash the firmware from the board and adb to tunneling flash command from the host PC.`,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			feedback.SetQuiet(quite)
			feedback.SetVerbose(verbose)
		},
	}
	root.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	root.PersistentFlags().BoolVar(&quite, "quite", false, "Enable quite logging (overrides verbose)")

	root.AddCommand(newVersionCmd())
	root.AddCommand(newUploadCmd())

	return root
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of remoteocd",
		Run: func(cmd *cobra.Command, args []string) {
			feedback.Printf("remoteocd version: %s", Version)
		},
	}
}

func newUploadCmd() *cobra.Command {
	var serial string
	var adbPath string
	var password string
	var address string
	var files []string
	upload := &cobra.Command{
		Use:   "upload <binary> <loader>",
		Args:  cobra.ExactArgs(2),
		Short: "Run a recipe for a specific board",
		RunE: func(cmd *cobra.Command, args []string) error {
			binaryPath := paths.New(args[0])
			if !binaryPath.Exist() {
				return fmt.Errorf("file %q does not exist", binaryPath.String())
			}
			loaderPath := paths.New(args[1])
			if !loaderPath.Exist() {
				return fmt.Errorf("file %q does not exist", loaderPath.String())
			}

			var filesPaths paths.PathList
			for _, f := range files {
				p := paths.New(f)
				if !p.Exist() {
					return fmt.Errorf("openocd configuration file %q does not exist", f)
				}
				filesPaths = append(filesPaths, p)
			}

			cmd.SilenceUsage = true // Do not print usage on error.

			var cmder board.Boarder
			if board.OnBoard {
				cmder = &board.LocalCmd{}
			} else {
				switch {
				case serial != "":
					cmder = &board.ADBCmd{
						Serial:  serial,
						ADBPath: adbPath,
					}
				case address != "":
					var err error
					cmder, err = board.NewSSHCmd(password, address)
					if err != nil {
						return fmt.Errorf("failed to create SSH connection: %w", err)
					}
				default:
					return fmt.Errorf("either --serial or --address must be provided when not running on the board")
				}
			}

			return flash(cmd.Context(), cmder, binaryPath, loaderPath, filesPaths)
		},
	}
	upload.Flags().StringVar(&adbPath, "adb-path", "", "Path to adb binary, if not set it will try to find it")
	upload.Flags().StringVarP(&serial, "serial", "s", "", "USB serial number of the connected board")
	upload.Flags().StringVarP(&address, "address", "a", "", "SSH address of the remote host")
	upload.Flags().StringVarP(&password, "password", "p", "", "Password for the SSH connection")
	upload.Flags().StringArrayVarP(&files, "file", "f", []string{}, "openocd configuration files to use")

	return upload
}
