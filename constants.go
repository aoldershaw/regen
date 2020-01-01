package regen

import "strings"

// Flag represents one or more flags applied to a group. Multiple flags can be applied using
// a bitwise OR |. Refer to: https://golang.org/pkg/regexp/syntax/
type Flag uint

const (
	// FlagCaseInsensitive corresponds with flag "i" (case-insensitive)
	FlagCaseInsensitive Flag = 1 << iota
	// FlagMultiLine corresponds with flag "m" (multi-line mode: ^ and $ match begin/end line in addition to begin/end text)
	FlagMultiLine
	// FlagMatchNewLine corresponds with flag "s" (let . match \n)
	FlagMatchNewLine
	// FlagUngreedy corresponds with flag "U" (ungreedy: swap meaning of x* and x*?, x+ and x+?, etc)
	FlagUngreedy
)

// String gives the flags' string representation (if all the flags were to be set to true)
func (f Flag) String() string {
	var sb strings.Builder
	if f&FlagCaseInsensitive != 0 {
		sb.WriteByte('i')
	}
	if f&FlagMultiLine != 0 {
		sb.WriteByte('m')
	}
	if f&FlagMatchNewLine != 0 {
		sb.WriteByte('s')
	}
	if f&FlagUngreedy != 0 {
		sb.WriteByte('U')
	}
	return sb.String()
}

var (
	LineStart        = Raw(`^`)
	LineEnd          = Raw(`$`)
	TextStart        = Raw(`\A`)
	TextEnd          = Raw(`\z`)
	ASCIIBoundary    = Raw(`\b`)
	NotASCIIBoundary = Raw(`\B`)

	Any              = Raw(`.`)
	Digit            = perlCharClass('d')
	Whitespace       = perlCharClass('s')
	WordCharacter    = perlCharClass('w')
)
