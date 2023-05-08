package compiler

type TokenType int

//go:generate go run golang.org/x/tools/cmd/stringer -type=TokenType -output=token_types_string.go
const (
	TokenTypeIdentifier TokenType = iota + 1

	// Keywords
	TokenTypeLet
	TokenTypeConst
	TokenTypeFunc
	TokenTypeIf
	TokenTypeElse
	TokenTypeFor
	TokenTypeLoop
	TokenTypeReturn
	TokenTypeBreak
	TokenTypeContinue
	TokenTypeStruct
	TokenTypeInterface

	// Punctuations
	TokenTypeLineBreak             // \n
	TokenTypeSemi                  // ;
	TokenTypeComma                 // ,
	TokenTypeColon                 // :
	TokenTypeLeftParen             // (
	TokenTypeRightParen            // )
	TokenTypeLeftBrace             // {
	TokenTypeRightBrace            // }
	TokenTypeLeftBracket           // [
	TokenTypeRightBracket          // ]
	TokenTypeDot                   // .
	TokenTypeEqual                 // =
	TokenTypeDoubleEqual           // ==
	TokenTypeBangEqual             // !=
	TokenTypePlus                  // +
	TokenTypeMinus                 // -
	TokenTypeStar                  // *
	TokenTypeDoubleStar            // **
	TokenTypeSlash                 // /
	TokenTypePercent               // %
	TokenTypeAlpha                 // @
	TokenTypeWavy                  // ~
	TokenTypeCaret                 // ^
	TokenTypeAmpersand             // &
	TokenTypeBang                  // !
	TokenTypeVertical              // |
	TokenTypeLeftAngle             // <
	TokenTypeRightAngle            // >
	TokenTypeDoubleLeftAngle       // <<
	TokenTypeDoubleRightAngle      // >>
	TokenTypeDoubleAmpersand       // &&
	TokenTypeDoubleVertical        // ||
	TokenTypeLeftAngleEqual        // <=
	TokenTypeRightAngleEqual       // >=
	TokenTypeArrow                 // =>
	TokenTypeDoublePlus            // ++
	TokenTypeDoubleMinus           // --
	TokenTypePlusEqual             // +=
	TokenTypeMinusEqual            // -=
	TokenTypeStarEqual             // *=
	TokenTypeSlashEqual            // /=
	TokenTypePercentEqual          // %=
	TokenTypeDoubleLeftAngleEqual  // <<=
	TokenTypeDoubleRightAngleEqual // >>=
	TokenTypeAmpersandEqual        // &=
	TokenTypeVerticalEqual         // |=
	TokenTypeCaretEqual            // ^=
	TokenTypeEllipsis              // ...
	TokenTypeDoubleDots            // ..
	TokenTypeQuestion              // ?
	TokenTypeQuestionDot           // ?.
	TokenTypeDoubleQuestion        // ??
	TokenTypeInterplolationStart   // ${

	// Literals
	TokenTypeDecimalInteger
	TokenTypeOctalInteger
	TokenTypeHexadecimalInteger
	TokenTypeBinaryInteger
	TokenTypeExponent
	TokenTypeFloat
	TokenTypeRune
	TokenTypeString
	TokenTypeTemplateString
	TokenTypeTrue
	TokenTypeFalse

	TokenTypeLineComment
)

var KeywordTokenMap = map[string]TokenType{
	"let":       TokenTypeLet,
	"const":     TokenTypeConst,
	"func":      TokenTypeFunc,
	"if":        TokenTypeIf,
	"else":      TokenTypeElse,
	"for":       TokenTypeFor,
	"loop":      TokenTypeLoop,
	"return":    TokenTypeReturn,
	"break":     TokenTypeBreak,
	"continue":  TokenTypeContinue,
	"struct":    TokenTypeStruct,
	"interface": TokenTypeInterface,
	"true":      TokenTypeTrue,
	"false":     TokenTypeFalse,
}

func isKeyword(s string) (TokenType, bool) {
	t, ok := KeywordTokenMap[s]
	return t, ok
}

// Token is the smallest unit of source code.
// It is used to represent a word, a number, a string, etc.
type Token struct {
	Type    TokenType
	Pos     *Position
	Content string
}
