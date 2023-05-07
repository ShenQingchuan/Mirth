//go:build tools
// +build tools

package main

// This file imports things required by build scripts,
// to force `go mod` to see them as dependencies
import (
	_ "golang.org/x/tools/cmd/stringer"
)
