package compiler

import (
	"fmt"
	"mirth/shared"

	"github.com/fatih/color"
)

// Position represents a position in the source code.
// It is used for error reporting.
type Position struct {
	Offset int
	Line   int
	Column int
}

func (p *Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

func CreatePositon(offset, line, colum int) *Position {
	return &Position{offset, line, colum}
}

type DiagnosticType int
type DiagnosticCode int

const (
	DiagnosticError DiagnosticType = iota
	DiagnosticWarning
)

// Diagnostic codes:
const (
	// ---- 1. Error Codes:
	// UnknownError is an fallback error code for errors that don't have a clear specification.
	UnknownError DiagnosticCode = iota

	// Scanner errors, mostly related to syntax issues
	UnexpectedToken
	FailedToRetrieveToken

	// ---- 2. Warning Codes:
	// UnknownWarning is an fallback warning code for warnings that don't have a clear specification.
	UnknownWarning
)

// Error type represents something unexpected in the source code.
type Diagnostic struct {
	Type DiagnosticType
	Code DiagnosticCode
	Pos  *Position
	Msg  string
}

func (d *Diagnostic) String() string {
	colorCodes := []color.Attribute{
		color.FgWhite,
		shared.Ternary(d.Type == DiagnosticError, color.BgRed, color.BgYellow),
		color.Bold,
	}

	return fmt.Sprintf("%s %s: %s", shared.ColorString(
		shared.Ternary(d.Type == DiagnosticError, " Error ", " Warning "),
		colorCodes,
	), d.Pos, d.Msg)
}
func (d *Diagnostic) Error() string {
	return d.String()
}

func CreateErrorDiagnostic(code DiagnosticCode, pos *Position, msg string) *Diagnostic {
	return &Diagnostic{DiagnosticError, code, pos, msg}
}
func CreateWarningDiagnostic(code DiagnosticCode, pos *Position, msg string) *Diagnostic {
	return &Diagnostic{DiagnosticWarning, code, pos, msg}
}
