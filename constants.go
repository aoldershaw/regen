package regen

import "strings"

type Flag uint

const (
	FlagCaseInsensitive Flag = 1 << iota
	FlagMultiLine
	FlagMatchNewLine
	FlagUngreedy
)

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
	Digit            = Raw(`\d`)
	NotDigit         = Raw(`\D`)
	Whitespace       = Raw(`\s`)
	NotWhitespace    = Raw(`\S`)
	WordCharacter    = Raw(`\w`)
	NotWordCharacter = Raw(`\W`)
)
