package regen

import (
	"regexp"
	"strconv"
	"strings"
)

// TODO: godocs...

type Regexp interface {
	Regexp() string
	Group() GroupedRegexp
	Repeat() RepeatedRegexp
	Optional() Regexp
}

type CharClass interface {
	Regexp
	Negate() CharClass
	Append(classes ...CharClass) CharClass
	charSetRegexp() string
}

type GroupedRegexp interface {
	Regexp
	Capture() GroupedRegexp
	CaptureAs(name string) GroupedRegexp
	NoCapture() GroupedRegexp
	SetFlags(flags Flag) GroupedRegexp
	UnsetFlags(flags Flag) GroupedRegexp
}

type RepeatedRegexp interface {
	Regexp
	Min(uint) RepeatedRegexp
	Max(uint) RepeatedRegexp
	Exactly(uint) RepeatedRegexp
	Greedy() RepeatedRegexp
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

func OneOf(choices ...Regexp) Regexp {
	return groupedRegexp{
		re: multiRegexp{
			res:       choices,
			separator: "|",
		},
	}
}

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

func Raw(s string) Regexp {
	return literalRegexp{
		re: s,
	}
}

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
