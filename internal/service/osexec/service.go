// meta1v is a command-line tool for viewing and manipulating metadata for Canon EOS-1V files of the EFD format.
// Copyright (C) 2026  Matt F
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

//go:generate mockgen -destination=./mocks/service_mock.go -package=osexec_test github.com/ma-tf/meta1v/internal/service/osexec Command,LookPath

// Package osexec provides an abstraction layer over os/exec for command execution.
//
// This package enables dependency injection and mocking of process execution for testing.
package osexec

import "os/exec"

// Command represents an external command that can be started and waited on.
type Command interface {
	// Start begins execution of the command without waiting for completion.
	Start() error

	// Wait waits for the command to complete and collects its exit status.
	Wait() error
}

type command struct {
	*exec.Cmd
}

// NewCommand wraps an exec.Cmd to implement the Command interface.
func NewCommand(cmd *exec.Cmd) Command {
	return &command{cmd}
}

// Start begins execution of the command without waiting for completion.
//
//nolint:wrapcheck // exec package errors are sufficient
func (c *command) Start() error {
	return c.Cmd.Start()
}

// Wait waits for the command to complete and collects its exit status.
//
//nolint:wrapcheck // exec package errors are sufficient
func (c *command) Wait() error {
	return c.Cmd.Wait()
}

// LookPath searches for an executable in the system PATH.
type LookPath interface {
	// LookPath searches for an executable named file in the directories
	// listed in the PATH environment variable.
	LookPath(file string) (string, error)
}

type lookPath struct{}

// NewLookPath creates a LookPath that delegates to exec.LookPath.
func NewLookPath() LookPath {
	return &lookPath{}
}

// LookPath searches for an executable named file in the directories
// listed in the PATH environment variable.
//
//nolint:wrapcheck // exec package errors are sufficient
func (l *lookPath) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
