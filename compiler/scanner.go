package compiler

import (
	"fmt"
	"mirth/shared"
	"strings"

	"github.com/fatih/color"
	"github.com/rivo/uniseg"
)

type Scanner struct {
	source []byte   // Buffer of source code
	lines  []string // Seperated lines of source code
	line   int      // Current line number
	column int      // Current column number
	offset int      // Offset of byte in source code

	// Cache for peeking
	currentRune *UniRune
	nextRune    *UniRune
}

type ScanResult = shared.Result[*Token, *Diagnostic]

type AvailableSource interface {
	string | []byte
}

func CreateScanner[S AvailableSource](source S) *Scanner {
	scanner := &Scanner{
		source: []byte(source),
		lines:  strings.Split(string(source), "\n"),
		line:   1,
		column: 1,
		offset: 0,
	}
	scanner.updatePeekCache()
	return scanner
}

func (s *Scanner) getCurrentPosition() *Position {
	return CreatePositon(
		s.offset,
		s.line,
		s.column,
	)
}

func (s *Scanner) peekRune() *UniRune {
	raw, _, _, _ := uniseg.FirstGraphemeCluster(s.source[s.offset:], -1)

	return &UniRune{string(raw), len(raw)}
}

func (s *Scanner) peekNextRune() *UniRune {
	currentLength := s.currentRune.byteLength
	raw, _, _, _ := uniseg.FirstGraphemeCluster(s.source[s.offset+currentLength:], -1)
	return &UniRune{string(raw), len(raw)}
}

func (s *Scanner) updatePeekCache() {
	s.currentRune = s.peekRune()
	s.nextRune = s.peekNextRune()
}

func (s *Scanner) peekForwardStepRune(step int) *UniRune {
	// `step - 1` is because we initialize forwardOffset with the first rune's byte length,
	// and `step` means the number of runes we want to peek forward.
	forwardOffset := s.currentRune.byteLength + step - 1
	contentSlice := s.source[s.offset+forwardOffset:]
	unisegState := -1
	var graphemeBytes []byte
	for i := 0; i < step-1; i++ {
		graphemeBytes, contentSlice, _, unisegState = uniseg.FirstGraphemeCluster(contentSlice, unisegState)
		if len(graphemeBytes) == 0 {
			break
		}
	}
	raw := string(graphemeBytes)
	return &UniRune{string(raw), len(graphemeBytes)}
}

func (s *Scanner) advanceRune() {
	s.offset += s.currentRune.byteLength
	s.updatePeekCache()
}

func (s *Scanner) advanceRuneByStep(step int) {
	forwardOffset := s.currentRune.byteLength
	unisegState := -1
	contentSlice := s.source[s.offset+forwardOffset:]
	var graphemeBytes []byte
	for i := 0; i < step-1; i++ {
		graphemeBytes, contentSlice, _, unisegState = uniseg.FirstGraphemeCluster(contentSlice, unisegState)
		if len(graphemeBytes) == 0 {
			break
		}
		forwardOffset += len(graphemeBytes)
	}
	s.offset += forwardOffset
	s.updatePeekCache()
}

func (s *Scanner) makeToken(tokenType TokenType, value string) *Token {
	s.column += uniseg.GraphemeClusterCount(value)
	return &Token{
		tokenType,
		CreatePositon(
			s.offset,
			s.line,
			s.column,
		),
		value,
	}
}

func (s *Scanner) createScannerErr(errCode DiagnosticCode, message string) *Diagnostic {
	return CreateErrorDiagnostic(
		errCode,
		CreatePositon(
			s.offset,
			s.line,
			s.column,
		),
		message,
	)
}

func (s *Scanner) createScannerWarn(warnCode DiagnosticCode, message string) *Diagnostic {
	return CreateWarningDiagnostic(
		warnCode,
		CreatePositon(
			s.offset,
			s.line,
			s.column,
		),
		message,
	)
}

func (s *Scanner) ResultOk(value *Token) *ScanResult {
	return &ScanResult{Value: value, Ok: true}
}

func (s *Scanner) ResultErr(err *Diagnostic) *ScanResult {
	return &ScanResult{Err: err}
}

func (s *Scanner) createScanResultErr(errCode DiagnosticCode, message string) *ScanResult {
	err := s.createScannerErr(errCode, message)
	return s.ResultErr(err)
}

func (s *Scanner) readLineComment() *ScanResult {
	var comment string
	for r := s.currentRune; !isLineBreak(r); r = s.currentRune {
		s.advanceRune()
		comment += r.raw
	}
	return s.ResultOk(s.makeToken(TokenTypeLineComment, comment))
}

func (s *Scanner) readIdentifier() *ScanResult {
	var identifier string
	for r := s.currentRune; s.offset < len(s.source) && isValidIdentifierRune(r); r = s.currentRune {
		s.advanceRune()
		identifier += r.raw
	}
	tokenType := TokenTypeIdentifier

	// Check if the identifier is a keyword
	if keywordTokenType, isKeywordToken := isKeyword(identifier); isKeywordToken {
		tokenType = keywordTokenType
	}
	return s.ResultOk(s.makeToken(tokenType, identifier))
}

func (s *Scanner) readNumber() *ScanResult {
	startFromZero := s.currentRune.isRune('0')
	hasDot := false
	hasExponent := false
	hasMultipleLeadingZero := false
	numberTokenType := shared.Ternary(
		startFromZero,
		TokenTypeOctalInteger,
		TokenTypeDecimalInteger,
	)
	numberRawStr := string(s.currentRune.raw)
	s.advanceRune() // Moving over the first zero

	// If there are multiple leading zeros for this number, should save only one.
	if startFromZero {
		for s.currentRune.isRune('0') {
			hasMultipleLeadingZero = true
			s.advanceRune()
		}
	}
	if hasMultipleLeadingZero && (isRadixSymbol(s.currentRune) ||
		s.currentRune.isRune('.') ||
		s.currentRune.isRune('e')) {
		return s.createScanResultErr(
			UnexpectedToken,
			"Unexpected token: multiple leading zeros before radix symbol",
		)
	}

	checkValidDigit := isDecimalDigit
	// Select the corresponding check function based on the radix symbol.
	if isRadixSymbol(s.currentRune) {
		switch s.currentRune.firstRune() {
		case 'b', 'B':
			checkValidDigit = isBinaryDigit
			numberTokenType = TokenTypeBinaryInteger
		case 'o', 'O':
			checkValidDigit = isOctalDigit
			numberTokenType = TokenTypeOctalInteger
		case 'x', 'X':
			checkValidDigit = isHexDigit
			numberTokenType = TokenTypeHexadecimalInteger
		}
		numberRawStr += string(s.currentRune.raw)
		s.advanceRune() // Moving over the radix symbol
	}

	for {
		if s.currentRune.isRune('.') {
			// If here're actually two or three dots, it's regard as range operator.
			if s.nextRune.isRune('.') {
				break // Quit the loop, and make token using the retrieved number
			}

			if !hasDot {
				if !hasExponent {
					hasDot = true
					numberTokenType = TokenTypeFloat
					numberRawStr += string(s.currentRune.raw)
					s.advanceRune()
					continue
				} else {
					return s.createScanResultErr(
						UnexpectedToken,
						"Unexpected token: decimal point '.' after exponent",
					)
				}
			} else {
				return s.createScanResultErr(
					UnexpectedToken,
					"Unexpected token: multiple decimal point '.'",
				)
			}
		}

		if s.currentRune.isRune('e') {
			if numberTokenType == TokenTypeHexadecimalInteger {
				s.advanceRune()
			} else if []rune(numberRawStr)[len(numberRawStr)-1] == '.' {
				return s.createScanResultErr(
					UnexpectedToken,
					"Unexpected token: exponent symbol 'e' after decimal point '.'",
				)
			} else if !hasExponent {
				hasExponent = true
				numberTokenType = TokenTypeExponent
				numberRawStr += string(s.currentRune.raw)
				s.advanceRune()

				// If there's '+' or '-' after 'e', it's a valid symbol, read it as well.
				if s.currentRune.isRune('+') || s.currentRune.isRune('-') {
					numberRawStr += string(s.currentRune.raw)
					s.advanceRune()
				}
			} else {
				return s.createScanResultErr(
					UnexpectedToken,
					"Unexpected token: multiple exponent symbol 'e'",
				)
			}
		} else if checkValidDigit(s.currentRune) {
			numberRawStr += string(s.currentRune.raw)
			s.advanceRune()
		} else {
			break
		}
	}

	// If the number is a exponent but starts with '0[oO]' or '0[bB]', it's invalid.
	runeListOfResult := []rune(numberRawStr)
	tipsColorfulPrefix := shared.ColorString(
		"Tips: ",
		[]color.Attribute{color.FgCyan, color.Bold},
	)
	if hasExponent && startFromZero && len(runeListOfResult) > 2 && isRadixSymbolRune(runeListOfResult[1]) {
		return s.createScanResultErr(
			UnexpectedToken,
			"Unexpected token: invalid number literal.\n"+tipsColorfulPrefix+"exponent should not start with '0o' or '0b'.",
		)
	}
	// If the last rune of the number is 'e', it's invalid.
	if runeListOfResult[len(runeListOfResult)-1] == 'e' {
		return s.createScanResultErr(
			UnexpectedToken,
			"Unexpected token: invalid number literal.\n"+tipsColorfulPrefix+"exponent should not be empty.",
		)
	}

	return s.ResultOk(
		s.makeToken(numberTokenType, numberRawStr),
	)
}

func (s *Scanner) readHexSequenceStrForRune(length int) *shared.Result[string, *Diagnostic] {
	unicodePointString := ""
	for i := 0; i < length; i++ {
		if !isHexDigit(s.currentRune) {
			return shared.ResultErr[string](
				CreateErrorDiagnostic(
					UnexpectedToken,
					s.getCurrentPosition(),
					fmt.Sprintf(
						"Unexpected token: invalid hexadecimal digit '%s' in rune escape sequence",
						string(s.currentRune.raw),
					),
				),
			)
		}
		unicodePointString += string(s.currentRune.raw)
		s.advanceRune()
	}
	// Convert the hexadecimal string to Golang rune
	strFromUnicodePoint := shared.UnicodePointToString(unicodePointString)
	if !strFromUnicodePoint.Ok {
		return shared.ResultErr[string](
			CreateErrorDiagnostic(
				UnexpectedToken,
				s.getCurrentPosition(),
				strFromUnicodePoint.Err.Error(),
			),
		)
	}
	return shared.ResultOk[string, *Diagnostic](
		strFromUnicodePoint.Unwrap(),
	)
}

func (s *Scanner) readRune() *ScanResult {
	// Moving over the first quote
	s.advanceRune()

	var runeContent string
	for !s.currentRune.isRune('\'') {
		if s.currentRune.isRune('\n') {
			return s.createScanResultErr(
				UnexpectedToken,
				"Unexpected token: newline is not allowed in character literal",
			)
		}

		if s.currentRune.isRune('\\') {
			if escaped, isSingleEscape := singleEscapeSymbolsRuneMap[s.nextRune.raw]; isSingleEscape {
				runeContent += escaped
				s.advanceRuneByStep(2) // Moving over the '\' and the escaped symbol
				continue
			}

			switch s.nextRune.raw {
			case "x":
				s.advanceRuneByStep(2) // Moving over the '\x'
				hexSeqStrResult := s.readHexSequenceStrForRune(2)
				if !hexSeqStrResult.Ok {
					return s.createScanResultErr(
						hexSeqStrResult.Err.Code,
						hexSeqStrResult.Err.Msg,
					)
				}
				runeContent += hexSeqStrResult.Unwrap()
			case "u":
				s.advanceRuneByStep(2) // Moving over the '\u'
				hexSeqStrResult := s.readHexSequenceStrForRune(4)
				if !hexSeqStrResult.Ok {
					return s.createScanResultErr(
						hexSeqStrResult.Err.Code,
						hexSeqStrResult.Err.Msg,
					)
				}
				runeContent += hexSeqStrResult.Unwrap()
			case "U":
				s.advanceRuneByStep(2) // Moving over the '\U'

				// Digits after \U must start with 0
				if !s.currentRune.isRune('0') {
					return s.createScanResultErr(
						UnexpectedToken,
						"Unexpected token: invalid first hexadecimal digit after '\\U' in rune escape sequence. Digits after '\\U' must start with 0",
					)
				}

				hexSeqStrResult := s.readHexSequenceStrForRune(8)
				if !hexSeqStrResult.Ok {
					return s.createScanResultErr(
						hexSeqStrResult.Err.Code,
						hexSeqStrResult.Err.Msg,
					)
				}
				runeContent += hexSeqStrResult.Unwrap()
			default:
				return s.createScanResultErr(
					UnexpectedToken,
					fmt.Sprintf(
						"Unexpected token: invalid escape symbol '%s'",
						string(s.nextRune.raw),
					),
				)
			}
		} else {
			runeContent += string(s.currentRune.raw)
			s.advanceRune()
		}
	}

	// Moving over the last quote
	s.advanceRune()
	return s.ResultOk(
		s.makeToken(TokenTypeRune, runeContent),
	)
}

func (s *Scanner) resultSingleRuneToken(tokenType TokenType, tokenContent string) *ScanResult {
	s.advanceRune()
	return s.ResultOk(
		s.makeToken(tokenType, tokenContent),
	)
}

func (s *Scanner) resultMultiRuneToken(tokenType TokenType, tokenContent string) *ScanResult {
	s.advanceRuneByStep(uniseg.GraphemeClusterCount(tokenContent))
	return s.ResultOk(
		s.makeToken(tokenType, tokenContent),
	)
}

func (s *Scanner) getNextToken() *ScanResult {
	for s.offset < len(s.source) {
		r := s.currentRune
		switch r.raw {
		default:
			// All the number literals start with a decimal digit:
			// - Hexadecimal number literals start with 0 and with a 'x' or 'X' after it
			// - Binary number literals start with 0 and with a 'b' or 'B' after it
			// - Octal number literals start with just 0 or 0o, 0O
			// - Decimal number literals start with a digit from 1 to 9
			if isDecimalDigit(r) {
				return s.readNumber()
			}
			return s.readIdentifier()
		case " ", "\t", "\r":
			// Skip whitespaces
			// '\r' is put here because it's a part of '\r\n',
			// only '\r' is not considered as a line break.
			s.advanceRune()
		case "\n":
			// Line break is considered as a token,
			// because it's used to separate statements.
			return s.resultSingleRuneToken(TokenTypeLineBreak, r.raw)
		case ";":
			return s.resultSingleRuneToken(TokenTypeSemi, r.raw)
		case ",":
			return s.resultSingleRuneToken(TokenTypeComma, r.raw)
		case ":":
			return s.resultSingleRuneToken(TokenTypeColon, r.raw)
		case "(":
			return s.resultSingleRuneToken(TokenTypeLeftParen, r.raw)
		case ")":
			return s.resultSingleRuneToken(TokenTypeRightParen, r.raw)
		case "{":
			return s.resultSingleRuneToken(TokenTypeLeftBrace, r.raw)
		case "}":
			return s.resultSingleRuneToken(TokenTypeRightBrace, r.raw)
		case "[":
			return s.resultSingleRuneToken(TokenTypeLeftBracket, r.raw)
		case "]":
			return s.resultSingleRuneToken(TokenTypeRightBracket, r.raw)
		case ".":
			// If here're actually two or three dots, it's regard as range operator.
			if s.nextRune.isRune('.') {
				if s.peekForwardStepRune(2).isRune('.') {
					return s.resultMultiRuneToken(TokenTypeEllipsis, "...")
				}
				return s.resultMultiRuneToken(TokenTypeDoubleDots, "..")
			}
			return s.resultSingleRuneToken(TokenTypeDot, r.raw)
		case "=":
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeDoubleEqual, "==")
			} else if s.nextRune.isRune('>') {
				return s.resultMultiRuneToken(TokenTypeArrow, "=>")
			}
			return s.resultSingleRuneToken(TokenTypeEqual, r.raw)
		case "+":
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypePlusEqual, "+=")
			} else if s.nextRune.isRune('+') {
				return s.resultMultiRuneToken(TokenTypeDoublePlus, "++")
			}
			return s.resultSingleRuneToken(TokenTypePlus, r.raw)
		case "-":
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeMinusEqual, "-=")
			} else if s.nextRune.isRune('-') {
				return s.resultMultiRuneToken(TokenTypeDoubleMinus, "--")
			}
			return s.resultSingleRuneToken(TokenTypeMinus, r.raw)
		case "*":
			// If here're actually two stars, it's regard as power operator.
			if s.nextRune.isRune('*') {
				return s.resultMultiRuneToken(TokenTypeDoubleStar, "**")
			}
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeStarEqual, "*=")
			}
			return s.resultSingleRuneToken(TokenTypeStar, r.raw)
		case "/":
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeSlashEqual, "/=")
			} else if s.nextRune.isRune('/') {
				return s.readLineComment()
			}
			return s.resultSingleRuneToken(TokenTypeSlash, r.raw)
		case "%":
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypePercentEqual, "%=")
			}
			return s.resultSingleRuneToken(TokenTypePercent, r.raw)
		case "&":
			if s.nextRune.isRune('&') {
				return s.resultMultiRuneToken(TokenTypeDoubleAmpersand, "&&")
			} else if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeAmpersandEqual, "&=")
			}
			return s.resultSingleRuneToken(TokenTypeAmpersand, r.raw)
		case "|":
			if s.nextRune.isRune('|') {
				return s.resultMultiRuneToken(TokenTypeDoubleVertical, "||")
			} else if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeVerticalEqual, "|=")
			}
			return s.resultSingleRuneToken(TokenTypeVertical, r.raw)
		case "^":
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeCaretEqual, "^=")
			}
			return s.resultSingleRuneToken(TokenTypeCaret, r.raw)
		case "~":
			return s.resultSingleRuneToken(TokenTypeWavy, r.raw)
		case "!":
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeBangEqual, "!=")
			}
			return s.resultSingleRuneToken(TokenTypeBang, r.raw)
		case "<":
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeLeftAngleEqual, "<=")
			} else if s.nextRune.isRune('<') {
				if s.peekForwardStepRune(2).isRune('=') {
					return s.resultMultiRuneToken(TokenTypeDoubleLeftAngleEqual, "<<=")
				}
				return s.resultMultiRuneToken(TokenTypeDoubleLeftAngle, "<<")
			}
			return s.resultSingleRuneToken(TokenTypeLeftAngle, r.raw)
		case ">":
			if s.nextRune.isRune('=') {
				return s.resultMultiRuneToken(TokenTypeRightAngleEqual, ">=")
			} else if s.nextRune.isRune('>') {
				if s.peekForwardStepRune(2).isRune('=') {
					return s.resultMultiRuneToken(TokenTypeDoubleRightAngleEqual, ">>=")
				}
				return s.resultMultiRuneToken(TokenTypeDoubleRightAngle, ">>")
			}
			return s.resultSingleRuneToken(TokenTypeRightAngle, r.raw)
		case "?":
			if s.nextRune.isRune('?') {
				return s.resultMultiRuneToken(TokenTypeDoubleQuestion, "??")
			} else if s.nextRune.isRune('.') {
				return s.resultMultiRuneToken(TokenTypeQuestionDot, "?.")
			}
			return s.resultSingleRuneToken(TokenTypeQuestion, r.raw)
		case "'":
			return s.readRune()
		}
	}

	return s.createScanResultErr(
		FailedToRetrieveToken,
		"Failed to retrieve next token.",
	)
}
