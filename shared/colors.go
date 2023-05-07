package shared

import (
	color "github.com/fatih/color"
)

type ColorCode int

// Color codes for terminal output.
const (
	White     ColorCode = 0
	Red       ColorCode = 31
	Green     ColorCode = 32
	Yellow    ColorCode = 33
	Blue      ColorCode = 34
	Magenta   ColorCode = 35
	Cyan      ColorCode = 36
	BgWhite   ColorCode = 40
	BgRed     ColorCode = 41
	BgGreen   ColorCode = 42
	BgYellow  ColorCode = 43
	BgBlue    ColorCode = 44
	BgMagenta ColorCode = 45
	BgCyan    ColorCode = 46
)

// Give a list of color codes and a string, return the string with the ASCII color applied.
func ColorString(str string, colorAttrs []color.Attribute) string {
	colorStr := color.New()
	for _, attr := range colorAttrs {
		colorStr.Add(attr)
	}
	return colorStr.Sprint(str)
}
