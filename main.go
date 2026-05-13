// This file is part of remoteocd.
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

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
		Use:   "upload <binary>...",
		Args:  cobra.MinimumNArgs(1),
		Short: "Run a recipe for a specific board",
		RunE: func(cmd *cobra.Command, args []string) error {
			binaries := make(paths.PathList, 0, len(args))
			for _, arg := range args {
				binaryPath := paths.New(arg)
				if !binaryPath.Exist() {
					return fmt.Errorf("file %q does not exist", binaryPath.String())
				}
				binaries.Add(binaryPath)
			}

			configs := make(paths.PathList, 0, len(files))
			for _, f := range files {
				p := paths.New(f)
				if !p.Exist() {
					return fmt.Errorf("openocd configuration file %q does not exist", f)
				}
				configs.Add(p)
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

			return flash(cmd.Context(), cmder, binaries, configs)
		},
	}
	upload.Flags().StringVar(&adbPath, "adb-path", "", "Path to adb binary, if not set it will try to find it")
	upload.Flags().StringVarP(&serial, "serial", "s", "", "USB serial number of the connected board")
	upload.Flags().StringVarP(&address, "address", "a", "", "SSH address of the remote host")
	upload.Flags().StringVarP(&password, "password", "p", "", "Password for the SSH connection")
	upload.Flags().StringArrayVarP(&files, "file", "f", []string{}, "openocd configuration files to use")

	return upload
}
