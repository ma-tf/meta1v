//go:generate mockgen -destination=./mocks/service_mock.go -package=osexec_test github.com/ma-tf/meta1v/internal/service/osexec Command,LookPath
package osexec

import "os/exec"

type Command interface {
	Start() error
	Wait() error
}

type command struct {
	*exec.Cmd
}

func NewCommand(cmd *exec.Cmd) Command {
	return &command{cmd}
}

//nolint:wrapcheck // exec package errors are sufficient
func (c *command) Start() error {
	return c.Cmd.Start()
}

//nolint:wrapcheck // exec package errors are sufficient
func (c *command) Wait() error {
	return c.Cmd.Wait()
}

type LookPath interface {
	LookPath(file string) (string, error)
}

type lookPath struct{}

func NewLookPath() LookPath {
	return &lookPath{}
}

//nolint:wrapcheck // exec package errors are sufficient
func (l *lookPath) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
