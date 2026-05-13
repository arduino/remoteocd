// This file is part of arduino-flasher-cli.
//
// Copyright 2025 ARDUINO SA (http://www.arduino.cc/)
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package board

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"

	"github.com/arduino/remoteocd/feedback"
)

var _ Boarder = (*SSHCmd)(nil)

const (
	user    = "arduino"
	sshPort = "22"
)

type SSHCmd struct {
	client *ssh.Client
}

func NewSSHCmd(password, address string) (*SSHCmd, error) {
	client, err := ssh.Dial("tcp", net.JoinHostPort(address, sshPort), &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		// TODO: audit the security of this setting
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // nolint:gosec
	})
	if err != nil {
		return nil, fmt.Errorf("failed to dial SSH: %w", err)
	}

	return &SSHCmd{client: client}, nil
}

func (s *SSHCmd) Run(ctx context.Context, args ...string) error {
	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stderr = feedback.GetStdout()
	session.Stdout = feedback.GetStdout()

	args = escapeArgs(args)

	return session.Run(strings.Join(args, " "))
}

func (c *SSHCmd) CopyTo(ctx context.Context, src, dst string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("faild to open file: %w", err)
	}
	defer f.Close()

	session.Stdin = f

	if err := session.Run(fmt.Sprintf("cat > %s", dst)); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (c *SSHCmd) MkDirAll(ctx context.Context, path string) error {
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if err := session.Run(fmt.Sprintf("mkdir -p %s", path)); err != nil {
		return fmt.Errorf("failt to make dir: %w", err)
	}
	return nil
}
