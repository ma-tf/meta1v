// Package main is the entry point for the meta1v CLI tool.
//
// meta1v is a command-line utility for interacting with Canon EFD (Electronic Film Data?) files,
// allowing users to view, export, and manipulate metadata recorded by Canon EOS film cameras.
package main

import "github.com/ma-tf/meta1v/cmd"

func main() {
	cmd.Execute()
}
