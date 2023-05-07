package shared

// Any is an interface that can be used to store any value.
type Any interface{}

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
