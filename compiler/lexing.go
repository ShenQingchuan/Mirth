package compiler

import "regexp"

type UniRune struct {
	raw        string // maybe single rune or multiple runes
	byteLength int    // length of rune in bytes
}

var identifierTerminatorRegExp = regexp.MustCompile(`[ \t\n;:,(){}\[\].=?!*/%^|&~><+\-'"]`)
var singleEscapeSymbolsRuneMap = map[string]string{
	"n":  "\n",
	"t":  "\t",
	"\\": "\\",
	"'":  "'",
	"\"": "\"",
	"r":  "\r",
	"a":  "\a",
	"b":  "\b",
	"f":  "\f",
	"v":  "\v",
	"0":  "\000",
}

func (r *UniRune) String() string {
	return string(r.raw)
}

func (r *UniRune) isRune(otherRune rune) bool {
	return r.hasOnlyOneRune() && r.firstRune() == otherRune
}

func (r *UniRune) hasOnlyOneRune() bool {
	return len([]rune(r.raw)) == 1
}

func (r *UniRune) firstRune() rune {
	return []rune(r.raw)[0]
}

func isDecimalDigit(r *UniRune) bool {
	if !r.hasOnlyOneRune() {
		return false
	}
	rawRune := r.firstRune()
	return rawRune >= '0' && rawRune <= '9'
}

func isHexDigit(r *UniRune) bool {
	if !r.hasOnlyOneRune() {
		return false
	}
	rawRune := r.firstRune()
	return isDecimalDigit(r) ||
		(rawRune >= 'a' && rawRune <= 'f') ||
		(rawRune >= 'A' && rawRune <= 'F')
}

func isOctalDigit(r *UniRune) bool {
	return r.hasOnlyOneRune() && r.firstRune() >= '0' && r.firstRune() <= '7'
}

func isBinaryDigit(r *UniRune) bool {
	return r.hasOnlyOneRune() && (r.firstRune() == '0' || r.firstRune() == '1')
}

func isLineBreak(r *UniRune) bool {
	return r.raw == "\n"
}

func isRadixSymbolRune(r rune) bool {
	switch r {
	case 'b', 'B', 'o', 'O', 'x', 'X':
		return true
	default:
		return false
	}
}

func isRadixSymbol(r *UniRune) bool {
	if !r.hasOnlyOneRune() {
		return false
	}
	return isRadixSymbolRune(r.firstRune())
}

func isValidIdentifierRune(r *UniRune) bool {
	return !identifierTerminatorRegExp.MatchString(r.raw)
}
