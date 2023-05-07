package compiler

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestScanNumber(t *testing.T) {
	expectPassCases := []struct {
		description string
		content     string
		tokenType   TokenType
	}{
		{"Test scan simple number", "123", TokenTypeDecimalInteger},
		{"Test scan hexadecimal", "0xBABEFACE", TokenTypeHexadecimalInteger},
		{"Test scan binary", "0b101010", TokenTypeBinaryInteger},
		{"Test scan octal", "0o777", TokenTypeOctalInteger},
		{"Test scan float", "123.456", TokenTypeFloat},
		{"Test scan simple exponent", "2e10", TokenTypeExponent},
		{"Test scan float with exponent", "123.456e10", TokenTypeExponent},
		{"Test scan float with positive exponent", "123.456e+10", TokenTypeExponent},
		{"Test scan float with negative exponent", "123.456e-10", TokenTypeExponent},
	}
	for _, testExpect := range expectPassCases {
		Convey(testExpect.description, t, func() {
			scanner := CreateScanner(testExpect.content)
			token := scanner.getNextToken().Unwrap()
			So(token.Type, ShouldEqual, testExpect.tokenType)
			So(token.Content, ShouldEqual, testExpect.content)
		})
	}

	expectFailCases := []struct {
		description string
		content     string
		errCode     DiagnosticCode
		errOffset   int
		errMsg      string
	}{
		{
			"Test scan octal exponent", "0o3e2",
			UnexpectedToken, 5,
			"Unexpected token: invalid number literal.\nTips: exponent should not start with '0o' or '0b'.",
		},
		{
			"Test empty exponent", "123e",
			UnexpectedToken, 4,
			"Unexpected token: invalid number literal.\nTips: exponent should not be empty.",
		},
		{
			"Test multiple dots in float", "123.456.789",
			UnexpectedToken, 7,
			"Unexpected token: multiple decimal point '.'",
		},
		{
			"Test dot after exponent", "123e.456",
			UnexpectedToken, 4,
			"Unexpected token: decimal point '.' after exponent",
		},
		{
			"Test exponent after dot", "123.e456",
			UnexpectedToken, 4,
			"Unexpected token: exponent symbol 'e' after decimal point '.'",
		},
		{
			"Test multiple exponent", "123e456e789",
			UnexpectedToken, 7,
			"Unexpected token: multiple exponent symbol 'e'",
		},
		{
			"Test multiple leading zeros before radix symbol", "000b101010",
			UnexpectedToken, 3,
			"Unexpected token: multiple leading zeros before radix symbol",
		},
	}
	for _, testExpect := range expectFailCases {
		Convey(testExpect.description, t, func() {
			scanner := CreateScanner(testExpect.content)
			scanResult := scanner.getNextToken()
			So(scanResult.Err, ShouldNotBeNil)
			So(scanResult.Err.Code, ShouldEqual, testExpect.errCode)
			So(scanResult.Err.Msg, ShouldEqual, testExpect.errMsg)
			So(scanResult.Err.Pos.Offset, ShouldEqual, testExpect.errOffset)
		})
	}
}

func TestScanPunctuations(t *testing.T) {
	expectPuncs := map[string]TokenType{
		"(":   TokenTypeLeftParen,
		")":   TokenTypeRightParen,
		"{":   TokenTypeLeftBrace,
		"}":   TokenTypeRightBrace,
		"[":   TokenTypeLeftBracket,
		"]":   TokenTypeRightBracket,
		",":   TokenTypeComma,
		":":   TokenTypeColon,
		";":   TokenTypeSemi,
		".":   TokenTypeDot,
		"..":  TokenTypeDoubleDots,
		"...": TokenTypeEllipsis,
		"+":   TokenTypePlus,
		"-":   TokenTypeMinus,
		"*":   TokenTypeStar,
		"/":   TokenTypeSlash,
		"%":   TokenTypePercent,
		"&":   TokenTypeAmpersand,
		"|":   TokenTypeVertical,
		"^":   TokenTypeCaret,
		"~":   TokenTypeWavy,
		"=":   TokenTypeEqual,
		"==":  TokenTypeDoubleEqual,
		"!=":  TokenTypeBangEqual,
		"<":   TokenTypeLeftAngle,
		"<=":  TokenTypeLeftAngleEqual,
		">":   TokenTypeRightAngle,
		">=":  TokenTypeRightAngleEqual,
		"!":   TokenTypeBang,
		"&&":  TokenTypeDoubleAmpersand,
		"||":  TokenTypeDoubleVertical,
		"<<":  TokenTypeDoubleLeftAngle,
		">>":  TokenTypeDoubleRightAngle,
		"+=":  TokenTypePlusEqual,
		"-=":  TokenTypeMinusEqual,
		"*=":  TokenTypeStarEqual,
		"/=":  TokenTypeSlashEqual,
		"%=":  TokenTypePercentEqual,
		"&=":  TokenTypeAmpersandEqual,
		"|=":  TokenTypeVerticalEqual,
		"^=":  TokenTypeCaretEqual,
		"<<=": TokenTypeDoubleLeftAngleEqual,
		">>=": TokenTypeDoubleRightAngleEqual,
		"++":  TokenTypeDoublePlus,
		"--":  TokenTypeDoubleMinus,
		"=>":  TokenTypeArrow,
		"?":   TokenTypeQuestion,
		"??":  TokenTypeDoubleQuestion,
		"?.":  TokenTypeQuestionDot,
	}

	Convey("Test scan punctuations", t, func() {
		for puncStr, puncTokenType := range expectPuncs {
			scanner := CreateScanner(puncStr)
			token := scanner.getNextToken().Unwrap()
			So(token.Type, ShouldEqual, puncTokenType)
			So(token.Content, ShouldEqual, puncStr)
		}
	})
}

func TestScanLineComment(t *testing.T) {
	source := "123 > 1.5e3\n// This is a line comment\n123"
	expectTokenTypes := []TokenType{
		TokenTypeDecimalInteger,
		TokenTypeRightAngle,
		TokenTypeExponent,
		TokenTypeLineBreak,
		TokenTypeLineComment,
		TokenTypeLineBreak,
		TokenTypeDecimalInteger,
	}
	scanner := CreateScanner(source)
	var tokenList []*Token
	// Get the token list
	for token := scanner.getNextToken(); token.Ok; token = scanner.getNextToken() {
		tokenList = append(tokenList, token.Unwrap())
	}

	Convey("Test scan line comment", t, func() {
		for i, tokenType := range expectTokenTypes {
			So(tokenList[i].Type, ShouldEqual, tokenType)
		}
	})
}
