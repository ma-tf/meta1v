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
