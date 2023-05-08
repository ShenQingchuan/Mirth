package shared

import (
	"fmt"
	"strconv"
)

// Ternary is a grammar sugar function for ternary operator in other languages.
func Ternary[T any](condition bool, forTrue, forFalse T) T {
	if condition {
		return forTrue
	} else {
		return forFalse
	}
}

// Get the suffix in English for a ordinal number.
func GetNumberSuffix(num int) string {
	if num == 0 {
		return ""
	} else if num == 1 {
		return "st"
	} else if num == 2 {
		return "nd"
	} else if num == 3 {
		return "rd"
	} else {
		return "th"
	}
}

func UnicodePointToString(unicodePointStr string) *Result[string, error] {
	unicodePoints, err := strconv.ParseInt(unicodePointStr, 16, 32)
	if err != nil {
		return ResultErr[string](
			fmt.Errorf("invalid unicode point digits: %s", err.Error()),
		)
	}
	return ResultOk[string, error](
		string(rune(unicodePoints)),
	)
}
