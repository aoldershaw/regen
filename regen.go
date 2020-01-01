package regen

import (
	"regexp"
	"strconv"
	"strings"
)

// Regexp is a representation of an uncompiled regular expression
type Regexp interface {
	// Regexp returns the regular expression as a string, which is commonly passed into regexp.MustCompile
	Regexp() string
	// Group returns a new Regexp that is in parentheses (if it is not already), making it a capturing group
	Group() GroupedRegexp
	// Repeat returns a new Regexp that is repeated, by default 0 to many times (equivalent to adding *).
	// This may wrap the regular expression in parentheses
	Repeat() RepeatedRegexp
	// Optional returns a new Regexp that can  appear 0 or 1 times (equivalent to adding ?).
	// This may wrap the regular expression in parentheses
	Optional() Regexp
}

// CharClass is a Regexp that represents a class of possible characters.
// This can be a set of allowed characters (e.g. [abc]), a range of allowed characters (e.g. [a-z]),
// a unicode character class (e.g. \p{Greek}), or an ASCII character class (e.g. [[:alpha:]])
type CharClass interface {
	Regexp
	// Negate returns a new CharClass that matches a character if and only if that character is
	// not matched in the original CharClass
	Negate() CharClass
	// Append returns a new CharClass that represents the union of the original CharClass, plus all
	// CharClasses in classes
	Append(classes ...CharClass) CharClass
	charSetRegexp() string
}

// GroupedRegexp is a Regexp that is wrapped in parentheses. It may or may not be a capturing group,
// and may have one or more flags enabled and/or disabled
type GroupedRegexp interface {
	Regexp
	// Capture returns a new GroupedRegexp that is a capturing group.
	// Since all GroupedRegexps are captured by default, this is only useful if you have an existing
	// non-capturing group Regexp that you wish to convert into a capturing group.
	Capture() GroupedRegexp
	// CaptureAs returns a new GroupedRegexp that is a named capturing group.
	// This name will appear in the SubexpNames function of a compiled *regexp.Regexp (from the standard library)
	CaptureAs(name string) GroupedRegexp
	// NoCapture returns a new GroupedRegexp that is non-capturing group
	NoCapture() GroupedRegexp
	// SetFlags returns a new GroupedRegexp that enables one or more Flags in the scope of the current group.
	// Multiple Flags can be passed in by joining them with the bitwise or |
	SetFlags(flags Flag) GroupedRegexp
	// UnsetFlags returns a new GroupedRegexp that explicitly disables one or more Flags in the scope of the current group.
	// Note that this is not always equivalent to not setting a flag. For instance, if a GroupedRegexp A is nested inside
	// another GroupedRegexp B, where B has a flag set, that flag will also apply to A (unless explicitly unset).
	// Multiple Flags can be passed in by joining them with the bitwise or |
	UnsetFlags(flags Flag) GroupedRegexp
}

// RepeatedRegexp is a Regexp that can be repeated some number of times
type RepeatedRegexp interface {
	Regexp
	// Min returns a new RepeatedRegexp that must appear at least min times
	Min(min uint) RepeatedRegexp
	// Max returns a new RepeatedRegexp that must appear at least max times
	Max(max uint) RepeatedRegexp
	// Exactly returns a new RepeatedRegexp that must appear exactly num times
	Exactly(num uint) RepeatedRegexp
	// Greedy returns a new RepeatedRegexp that prefers more matches.
	// Since all RepeatedRegexps are greedy by default, this is only useful if you have an existing
	// ungreedy RepeatedRegexp that you wish to convert into a greedy one.
	Greedy() RepeatedRegexp
	// Ungreedy returns a new RepeatedRegexp that prefers fewer matches.
	Ungreedy() RepeatedRegexp
}

type groupedRegexp struct {
	re         Regexp
	name       string
	setFlags   Flag
	unsetFlags Flag
	noCapture  bool
}

func (g groupedRegexp) Regexp() string {
	var sb strings.Builder
	sb.WriteByte('(')
	if g.name != "" {
		sb.WriteString("?P<")
		sb.WriteString(g.name)
		sb.WriteByte('>')
	}

	var flagsb strings.Builder
	if g.setFlags != 0 || g.unsetFlags != 0 {
		if g.setFlags != 0 {
			flagsb.WriteString(g.setFlags.String())
		}
		if g.unsetFlags != 0 {
			flagsb.WriteByte('-')
			flagsb.WriteString(g.unsetFlags.String())
		}
	}

	if g.noCapture {
		sb.WriteByte('?')
		sb.WriteString(flagsb.String())
		sb.WriteByte(':')
	} else if flagsb.Len() > 0 {
		sb.WriteString("(?")
		sb.WriteString(flagsb.String())
		sb.WriteString(")")
	}

	sb.WriteString(g.re.Regexp())
	sb.WriteByte(')')
	return sb.String()
}

func (g groupedRegexp) Group() GroupedRegexp {
	return g
}

func (g groupedRegexp) Repeat() RepeatedRegexp {
	return repeatedRegexp{re: g}
}

func (g groupedRegexp) Optional() Regexp {
	return repeatedRegexp{re: g}.Min(0).Max(1)
}

func (g groupedRegexp) Capture() GroupedRegexp {
	g.noCapture = false
	g.name = ""
	return g
}

func (g groupedRegexp) CaptureAs(name string) GroupedRegexp {
	g.noCapture = false
	g.name = name
	return g
}

func (g groupedRegexp) NoCapture() GroupedRegexp {
	g.noCapture = true
	g.name = ""
	return g
}

func (g groupedRegexp) SetFlags(flags Flag) GroupedRegexp {
	g.setFlags = flags
	return g
}

func (g groupedRegexp) UnsetFlags(flags Flag) GroupedRegexp {
	g.unsetFlags = flags
	return g
}

type repeatedRegexp struct {
	re       Regexp
	min      uint
	hasMin   bool
	max      uint
	hasMax   bool
	ungreedy bool
}

func (r repeatedRegexp) Regexp() string {
	subRe := r.re.Regexp()
	requiresParens := true
	if _, ok := r.re.(GroupedRegexp); ok {
		requiresParens = false
	}
	if _, ok := r.re.(CharClass); ok {
		requiresParens = false
	}
	if len(subRe) == 1 {
		requiresParens = false
	}
	if len(subRe) == 2 && subRe[0] == '\\' {
		requiresParens = false
	}
	var sb strings.Builder
	if requiresParens {
		sb.WriteByte('(')
	}
	sb.WriteString(subRe)
	if requiresParens {
		sb.WriteByte(')')
	}
	if !r.hasMax {
		if r.min == 0 {
			sb.WriteByte('*')
		} else if r.min == 1 {
			sb.WriteByte('+')
		} else {
			sb.WriteByte('{')
			sb.WriteString(strconv.Itoa(int(r.min)))
			sb.WriteString(",}")
		}
	} else {
		if r.max == 1 && r.min == 0 {
			sb.WriteByte('?')
		} else if r.min == r.max {
			sb.WriteByte('{')
			sb.WriteString(strconv.Itoa(int(r.min)))
			sb.WriteByte('}')
		} else {
			sb.WriteByte('{')
			sb.WriteString(strconv.Itoa(int(r.min)))
			sb.WriteByte(',')
			sb.WriteString(strconv.Itoa(int(r.max)))
			sb.WriteByte('}')
		}
	}
	if r.ungreedy {
		sb.WriteByte('?')
	}
	return sb.String()
}

func (r repeatedRegexp) Group() GroupedRegexp {
	return groupedRegexp{re: r}
}

func (r repeatedRegexp) Repeat() RepeatedRegexp {
	return repeatedRegexp{re: r}
}

func (r repeatedRegexp) Optional() Regexp {
	return repeatedRegexp{re: r}.Min(0).Max(1)
}

func (r repeatedRegexp) Min(min uint) RepeatedRegexp {
	r.min = min
	r.hasMin = true
	return r
}

func (r repeatedRegexp) Max(max uint) RepeatedRegexp {
	r.max = max
	r.hasMax = true
	return r
}

func (r repeatedRegexp) Exactly(num uint) RepeatedRegexp {
	r.min = num
	r.hasMin = true
	r.max = num
	r.hasMax = true
	return r
}

func (r repeatedRegexp) Greedy() RepeatedRegexp {
	r.ungreedy = false
	return r
}

func (r repeatedRegexp) Ungreedy() RepeatedRegexp {
	r.ungreedy = true
	return r
}

type multiRegexp struct {
	res       []Regexp
	separator string
}

// OneOf returns a new Regexp that matches any of choices, preferring the choices specified earlier
func OneOf(choices ...Regexp) Regexp {
	return groupedRegexp{
		re: multiRegexp{
			res:       choices,
			separator: "|",
		},
	}
}

// Sequence returns a new Regexp that expects each sub-Regexp to appear in order
func Sequence(subseqs ...Regexp) Regexp {
	return multiRegexp{
		res:       subseqs,
		separator: "",
	}
}

func (m multiRegexp) Regexp() string {
	var sb strings.Builder
	for i, re := range m.res {
		sb.WriteString(re.Regexp())
		if i < len(m.res)-1 {
			sb.WriteString(m.separator)
		}
	}
	return sb.String()
}

func (m multiRegexp) Group() GroupedRegexp {
	return groupedRegexp{re: m}
}

func (m multiRegexp) Repeat() RepeatedRegexp {
	return repeatedRegexp{re: m}
}

func (m multiRegexp) Optional() Regexp {
	return repeatedRegexp{re: m}.Min(0).Max(1)
}

type literalRegexp struct {
	re string
}

// Raw returns a Regexp that represents the literal regular expression string passed in.
// No validation is done on this string.
func Raw(s string) Regexp {
	return literalRegexp{
		re: s,
	}
}

// String returns a Regexp that matches the literal string.
// Regular expression metacharacters are escaped.
func String(s string) Regexp {
	return Raw(regexp.QuoteMeta(s))
}

func (l literalRegexp) Regexp() string {
	return l.re
}

func (l literalRegexp) Group() GroupedRegexp {
	return groupedRegexp{re: l}
}

func (l literalRegexp) Repeat() RepeatedRegexp {
	return repeatedRegexp{re: l}
}

func (l literalRegexp) Optional() Regexp {
	return repeatedRegexp{re: l}.Min(0).Max(1)
}

type charSetRegexp struct {
	chars       []rune
	charClasses []CharClass
	negated     bool
}

// CharSet returns a CharClass that matches any of the provided chars
func CharSet(chars ...rune) CharClass {
	return charSetRegexp{
		chars: chars,
	}
}

func (c charSetRegexp) Regexp() string {
	return "[" + c.charSetRegexp() + "]"
}

func (c charSetRegexp) Group() GroupedRegexp {
	return groupedRegexp{re: c}
}

func (c charSetRegexp) Repeat() RepeatedRegexp {
	return repeatedRegexp{re: c}
}

func (c charSetRegexp) Optional() Regexp {
	return repeatedRegexp{re: c}.Min(0).Max(1)
}

func writeCharSetRune(sb *strings.Builder, r rune) {
	if r == '\\' || r == '^' {
		sb.WriteByte('\\')
	}
	sb.WriteRune(r)
}

func (c charSetRegexp) charSetRegexp() string {
	var sb strings.Builder
	if c.negated {
		sb.WriteString("^")
	}
	for _, c := range c.chars {
		writeCharSetRune(&sb, c)
	}
	for _, c := range c.charClasses {
		sb.WriteString(c.charSetRegexp())
	}
	return sb.String()
}

func (c charSetRegexp) Append(classes ...CharClass) CharClass {
	return charSetRegexp{
		chars:       c.chars,
		charClasses: append(c.charClasses, classes...),
		negated:     c.negated,
	}
}

func (c charSetRegexp) Negate() CharClass {
	c.negated = !c.negated
	return c
}

type charRangeRegexp struct {
	start   rune
	end     rune
	negated bool
}

// CharRange returns a CharClass that matches any character between start and end, inclusive
func CharRange(start, end rune) CharClass {
	return charRangeRegexp{
		start: start,
		end:   end,
	}
}

func (c charRangeRegexp) Regexp() string {
	return "[" + c.charSetRegexp() + "]"
}

func (c charRangeRegexp) Group() GroupedRegexp {
	return groupedRegexp{re: c}
}

func (c charRangeRegexp) Repeat() RepeatedRegexp {
	return repeatedRegexp{re: c}
}

func (c charRangeRegexp) Optional() Regexp {
	return repeatedRegexp{re: c}.Min(0).Max(1)
}

func (c charRangeRegexp) charSetRegexp() string {
	var sb strings.Builder
	if c.negated {
		sb.WriteString("^")
	}
	writeCharSetRune(&sb, c.start)
	sb.WriteByte('-')
	writeCharSetRune(&sb, c.end)
	return sb.String()
}

func (c charRangeRegexp) Append(classes ...CharClass) CharClass {
	return charSetRegexp{
		charClasses: append([]CharClass{c}, classes...),
		negated:     c.negated,
	}
}

func (c charRangeRegexp) Negate() CharClass {
	c.negated = !c.negated
	return c
}

type asciiCharClassRegexp struct {
	name    string
	negated bool
}

// ASCIICharClass returns a CharClass that represents the set of ASCII characters specified by a name.
// For a list of available names, see https://golang.org/pkg/regexp/syntax/
func ASCIICharClass(name string) CharClass {
	return asciiCharClassRegexp{
		name: name,
	}
}

func (a asciiCharClassRegexp) Regexp() string {
	return "[" + a.charSetRegexp() + "]"
}

func (a asciiCharClassRegexp) Group() GroupedRegexp {
	return groupedRegexp{re: a}
}

func (a asciiCharClassRegexp) Repeat() RepeatedRegexp {
	return repeatedRegexp{re: a}
}

func (a asciiCharClassRegexp) Optional() Regexp {
	return repeatedRegexp{re: a}.Min(0).Max(1)
}

func (a asciiCharClassRegexp) charSetRegexp() string {
	negate := ""
	if a.negated {
		negate = "^"
	}
	return "[:" + negate + a.name + ":]"
}

func (a asciiCharClassRegexp) Append(classes ...CharClass) CharClass {
	return charSetRegexp{
		charClasses: append([]CharClass{a}, classes...),
		negated:     a.negated,
	}
}

func (a asciiCharClassRegexp) Negate() CharClass {
	a.negated = !a.negated
	return a
}

type unicodeCharClassRegexp struct {
	name    string
	negated bool
}

// UnicodeCharClass returns a CharClass that represents the set of Unicode characters specified by a name.
func UnicodeCharClass(name string) CharClass {
	return unicodeCharClassRegexp{
		name: name,
	}
}

func (u unicodeCharClassRegexp) Regexp() string {
	prefix := `\p`
	if u.negated {
		prefix = `\P`
	}
	name := u.name
	if len(name) > 1 {
		name = "{" + name + "}"
	}
	return prefix + name
}

func (u unicodeCharClassRegexp) Group() GroupedRegexp {
	return groupedRegexp{re: u}
}

func (u unicodeCharClassRegexp) Repeat() RepeatedRegexp {
	return repeatedRegexp{re: u}
}

func (u unicodeCharClassRegexp) Optional() Regexp {
	return repeatedRegexp{re: u}.Min(0).Max(1)
}

func (u unicodeCharClassRegexp) charSetRegexp() string {
	return u.Regexp()
}

func (u unicodeCharClassRegexp) Append(classes ...CharClass) CharClass {
	return charSetRegexp{
		charClasses: append([]CharClass{u}, classes...),
		negated:     u.negated,
	}
}

func (u unicodeCharClassRegexp) Negate() CharClass {
	u.negated = !u.negated
	return u
}
