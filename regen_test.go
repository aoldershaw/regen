package regen_test

import (
	"fmt"
	"github.com/aoldershaw/regen"
	"testing"
)

func TestRegen(t *testing.T) {
	tests := []struct {
		description string
		re          regen.Regexp
		expected    string
	}{
		{
			description: "OneOf joins regular expressions with '|'",
			re:          regen.OneOf(regen.String("a"), regen.String("bc")),
			expected:    "(a|bc)",
		},
		{
			description: "Raw returns the raw regexp",
			re:          regen.Raw(`\zhello\z`),
			expected:    `\zhello\z`,
		},
		{
			description: "String literals are escaped",
			re:          regen.String(`\zhello\z`),
			expected:    `\\zhello\\z`,
		},
		{
			description: "Sequence appends subsequences together",
			re:          regen.Sequence(regen.String("hello"), regen.String("world")),
			expected:    `helloworld`,
		},
		{
			description: "CharSet allows one of several runes",
			re:          regen.CharSet('h', 'こ', 'é'),
			expected:    `[hこé]`,
		},
		{
			description: "CharSet can be negated",
			re:          regen.CharSet('h', 'こ', 'é').Negate(),
			expected:    `[^hこé]`,
		},
		{
			description: "CharSet escapes special characters",
			re:          regen.CharSet('^', '\\'),
			expected:    `[\^\\]`,
		},
		{
			description: "CharRange allows two runes",
			re:          regen.CharRange('h', 'こ'),
			expected:    `[h-こ]`,
		},
		{
			description: "CharRange can be negated",
			re:          regen.CharRange('h', 'こ').Negate(),
			expected:    `[^h-こ]`,
		},
		{
			description: "CharRange escapes special characters",
			re:          regen.CharRange('\\', '^'),
			expected:    `[\\-\^]`,
		},
		{
			description: "ASCIICharClass takes a name",
			re:          regen.ASCIICharClass("alpha"),
			expected:    `[[:alpha:]]`,
		},
		{
			description: "ASCIICharClass can be negated",
			re:          regen.ASCIICharClass("alpha").Negate(),
			expected:    `[[:^alpha:]]`,
		},
		{
			description: "UnicodeCharClass takes a name",
			re:          regen.UnicodeCharClass("Greek"),
			expected:    `\p{Greek}`,
		},
		{
			description: "UnicodeCharClass can be negated",
			re:          regen.UnicodeCharClass("Greek").Negate(),
			expected:    `\P{Greek}`,
		},
		{
			description: "UnicodeCharClass does not wrap a single letter name in brackets",
			re:          regen.UnicodeCharClass("Z"),
			expected:    `\pZ`,
		},
		{
			description: "UnicodeCharClass does not wrap a single letter name in brackets",
			re:          regen.UnicodeCharClass("Z"),
			expected:    `\pZ`,
		},
		{
			description: "CharSet can be appended to with other CharClasses",
			re: regen.CharSet('h', 'こ', 'é').Append(
				regen.ASCIICharClass("alpha").Negate(),
				regen.UnicodeCharClass("Greek"),
				regen.CharSet('v'),
			),
			expected: `[hこé[:^alpha:]\p{Greek}v]`,
		},
		{
			description: "Grouping a Regexp",
			re:          regen.String("hello").Group(),
			expected:    `(hello)`,
		},
		{
			description: "Setting a Group to NoCapture",
			re:          regen.String("hello").Group().NoCapture(),
			expected:    `(?:hello)`,
		},
		{
			description: "Setting flags on a Group",
			re:          regen.String("hello").Group().SetFlags(regen.FlagCaseInsensitive | regen.FlagMultiLine),
			expected:    `((?im)hello)`,
		},
		{
			description: "Unsetting flags on a Group",
			re:          regen.String("hello").Group().UnsetFlags(regen.FlagCaseInsensitive | regen.FlagMultiLine),
			expected:    `((?-im)hello)`,
		},
		{
			description: "Setting and Unsetting flags on a Group",
			re:          regen.String("hello").Group().SetFlags(regen.FlagCaseInsensitive).UnsetFlags(regen.FlagMultiLine),
			expected:    `((?i-m)hello)`,
		},
		{
			description: "Setting flags on a NoCapture Group",
			re:          regen.String("hello").Group().NoCapture().SetFlags(regen.FlagCaseInsensitive | regen.FlagMultiLine),
			expected:    `(?im:hello)`,
		},
		{
			description: "Named capture group",
			re:          regen.String("hello").Group().CaptureAs("test"),
			expected:    `(?P<test>hello)`,
		},
		{
			description: "Named capture group with flags",
			re:          regen.String("hello").Group().CaptureAs("test").SetFlags(regen.FlagCaseInsensitive),
			expected:    `(?P<test>(?i)hello)`,
		},
		{
			description: "Repeating a Regexp defaults to 0 to many",
			re:          regen.String("hello").Repeat(),
			expected:    `(hello)*`,
		},
		{
			description: "Repeating with at least 1 uses +",
			re:          regen.String("hello").Repeat().Min(1),
			expected:    `(hello)+`,
		},
		{
			description: "Repeating with at least 2 uses {2,}",
			re:          regen.String("hello").Repeat().Min(2),
			expected:    `(hello){2,}`,
		},
		{
			description: "Repeating with max uses {0,max}",
			re:          regen.String("hello").Repeat().Max(10),
			expected:    `(hello){0,10}`,
		},
		{
			description: "Repeating with min and max uses {min,max}",
			re:          regen.String("hello").Repeat().Min(1).Max(10),
			expected:    `(hello){1,10}`,
		},
		{
			description: "Repeating with exactly num uses {num}",
			re:          regen.String("hello").Repeat().Exactly(5),
			expected:    `(hello){5}`,
		},
		{
			description: "Repeating with ungreedy appends a ?",
			re:          regen.String("hello").Repeat().Ungreedy(),
			expected:    `(hello)*?`,
		},
		{
			description: "Repeating a charset does not add parens",
			re:          regen.CharSet('h', 'e', 'y').Repeat(),
			expected:    `[hey]*`,
		},
		{
			description: "Repeating a grouped regexp does not add more parens",
			re:          regen.String("hello").Group().Repeat(),
			expected:    `(hello)*`,
		},
		{
			description: "Repeating a single character regexp does not add parens",
			re:          regen.String("h").Repeat(),
			expected:    `h*`,
		},
		{
			description: "Repeating a single escaped character regexp does not add parens",
			re:          regen.Raw(`\w`).Repeat(),
			expected:    `\w*`,
		},
		{
			description: "Nested repeats and groups",
			re:          regen.CharSet('h', 'e', 'y').Repeat().Min(1).Group().CaptureAs("test"),
			expected:    `(?P<test>[hey]+)`,
		},
		{
			description: "Optional regexp",
			re:          regen.CharSet('h', 'e', 'y').Optional(),
			expected:    `[hey]?`,
		},
		//{
		//	description: "Complex case",
		//	// (^\^?([^$\\]|\\[^z])*$)|(^\^?([^$\\]|\\[^z])*\.[*+](\$|\\z)$)
		//	re: regen.OneOf(
		//			regen.Sequence(
		//				regen.LineStart,
		//
		//			)
		//		),
		//	),
		//	expected: `[hこév[:^alpha:]\p{Greek}]`,
		//},
	}
	for _, tt := range tests {
		actual := tt.re.Regexp()
		if actual != tt.expected {
			t.Errorf(`regen test "%s" failed: got "%s", expected "%s"`, tt.description, actual, tt.expected)
		}
	}
}

func TestFlags(t *testing.T) {
	tests := []struct {
		description string
		flag        regen.Flag
		expected    string
	}{
		{
			description: "All flags",
			flag:        regen.FlagCaseInsensitive | regen.FlagMultiLine | regen.FlagMatchNewLine | regen.FlagUngreedy,
			expected:    "imsU",
		},
		{
			description: "Subset of flags",
			flag:        regen.FlagCaseInsensitive | regen.FlagMatchNewLine,
			expected:    "is",
		},
	}
	for _, tt := range tests {
		actual := tt.flag.String()
		if actual != tt.expected {
			t.Errorf(`flag test "%s" failed: got "%s", expected "%s"`, tt.description, actual, tt.expected)
		}
	}
}

func Example() {
	japaneseWord := regen.UnicodeCharClass("Hiragana").Append(
		regen.UnicodeCharClass("Katakana"),
		regen.UnicodeCharClass("Han"),
	).Repeat().Min(1)
	englishWord := regen.WordCharacter.Repeat().Min(1)
	re := regen.Sequence(
		regen.LineStart,
		regen.OneOf(japaneseWord, englishWord).Group().CaptureAs("greeting"),
		regen.Sequence(
			regen.Whitespace.Repeat().Min(1),
			regen.OneOf(regen.String("world"), regen.String("世界")),
		).Optional(),
		regen.LineEnd,
	)
	fmt.Println(re.Regexp())
	// Output: ^(?P<greeting>[\p{Hiragana}\p{Katakana}\p{Han}]+|\w+)(\s+(world|世界))?$
}
