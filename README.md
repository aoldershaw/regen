# regen

[![Documentation](https://godoc.org/github.com/aoldershaw/regen?status.svg)](http://godoc.org/github.com/aoldershaw/regen)

Regular Expression GENerator. It aims to provide a fluent syntax for constructing composable
Go regular expressions in a way that that is easy to read and write at the cost of being (much)
more verbose.

## Installation

```
$ go get github.com/aoldershaw/regen 
```

## Usage

```go
package main

import (
    "github.com/aoldershaw/regen"
    "regexp"
)

func main() {
    japaneseWord := regen.Union(
        regen.UnicodeCharClass("Hiragana"),
        regen.UnicodeCharClass("Katakana"),
        regen.UnicodeCharClass("Han"),
    ).Repeat().Min(1)

    englishWord := regen.WordCharacter.Repeat().Min(1)

    re := regexp.MustCompile(regen.Sequence(
        regen.LineStart,
        regen.OneOf(japaneseWord, englishWord).Group().CaptureAs("greeting"),
        regen.Sequence(
            regen.Whitespace.Repeat().Min(1),
            regen.OneOf(regen.String("world"), regen.String("世界")),
        ).Optional(),
        regen.LineEnd,
    ).Regexp())
    // Results in: ^(?P<greeting>[\p{Hiragana}\p{Katakana}\p{Han}]+|\w+)(\s+(world|世界))?$

    re.MatchString("こんにちは")   // == true (`greeting` capture group == "こんにちは")
    re.MatchString("hello 世界")  // == true (`greeting` capture group == "hello")
}
```

### Grouping/Capturing

You can create a capturing (or non-capturing) group by calling `.Group` on any regular expression.

```go
// Calling .Capture() is optional here, since the default is a capturing group
capture := regen.CharRange('A', 'Z').Group().Capture()
// Results in: ([A-Z])

namedCapture := regen.CharRange('A', 'Z').Group().CaptureAs("letter")
// Results in: (?P<letter>[A-Z])

noCapture := regen.CharRange('A', 'Z').Group().NoCapture()
// Results in: (?:[A-Z])
```

You can also set flags on a grouped regular expression:

```go
greeting := regen.String("hello").Group().SetFlags(regen.FlagCaseInsensitive | regen.FlagMultiLine)
// Results in: ((?im)hello)
```

Or you can unset flags if they were set in a top-level group:

```go
screamHello := regen.String("HELLO").Group().UnsetFlags(regen.CaseInsensitive)
greeting := regen.Sequence(
    screamHello,
    regen.Whitespace.Repeat().Min(1),
    regen.String("world")
).Group().SetFlags(regen.CaseInsensitive)
// Results in: ((?i)((?-i)HELLO)\s+world)

// greeting will match "HELLO world" or "HELLO WORLD", but not "hello world"
```

**Note:** not calling `.Group()` on a regular expression does not guarantee that it won't be grouped
in the resulting regexp. For instance, the result `regen.OneOf` is grouped. If you need to rely
on a particular ordering of capture groups, you should explicitly call `.Group().NoCapture()` on
sub expressions that should not be captured.

### Character Classes

There are several different types of character classes that are available:

* **`regen.CharSet`** - generate a whitelist or a blacklist of allowed characters
  * `regen.CharSet('a', 'e', 'i', 'o', 'u')` generates `[aeiou]`
  * `regen.CharSet('a', 'e', 'i', 'o', 'u').Negate()` generates `[^aeiou]`
* **`regen.CharRange`** - generate a range of characters to include/exclude
  * `regen.CharRange('a', 'z')` generates `[a-z]`
  * `regen.CharRange('a', 'z').Negate()` generates `[^a-z]`
* **`regen.ASCIICharClass`** - refer to a named ASCII character class to include/exclude.
  Refer to [regexp/syntax](https://golang.org/pkg/regexp/syntax/) for a list of available
  ASCII character classes.
  * `regen.ASCIICharClass("alpha")` generates `[[:alpha:]]`
  * `regen.ASCIICharClass("alpha").Negate()` generates `[[:^alpha:]]`
* **`regen.UnicodeCharClass`** - refer to a named Unicode character class to include/exclude.
  * `regen.UnicodeCharClass("Greek")` generates `\p{Greek}`
  * `regen.UnicodeCharClass("Greek").Negate()` generates `\P{Greek}`
* **`regen.Whitespace`, `regen.Digit`, `regen.WordCharacter`** - Perl character classes
  * `regen.Whitespace` generates `\s`, `regen.Whitespace.Negate()` generates `\S`
  * `regen.Digit` generates `\d`, `regen.Digit.Negate()` generates `\D`
  * `regen.WordCharacter` generates `\w`, `regen.WordCharacter.Negate()` generates `\W`

Multiple character classes can be joined using `regen.Union`:

```go
vowelsAndDigits := regen.Union(
    regen.CharSet('a', 'e', 'i', 'o', 'u'),
    regen.Digit,
)
// Results in: [aeiou\d]
```

### Raw Regular Expressions

If it is too awkward to construct your regular expression using this syntax,
but still want the other benefits of composition, use the `regen.Raw`
function to input raw regular expressions. For instance, to match base64 encoded
strings:

```go
re := regexp.MustCompile(regen.Sequence(
    regen.Raw(`[A-Za-z0-9+/]`).Repeat().Min(1),
    regen.String("=").Repeat().Max(2),
).Regexp())
// Results in: [A-Za-z0-9+/]+={0,2}
```

Note that this *could* be expressed without using `regen.Raw` as follows:

```go
re := regexp.MustCompile(regen.Sequence(
    regen.Union(
        regen.CharRange('A', 'Z'),
        regen.CharRange('a', 'z')
        regen.CharRange('0', '9'),
        regen.CharSet('+', '/'),
    ).Repeat().Min(1),
    regen.String("=").Repeat().Max(2),
).Regexp())
// Results in: [A-Za-z0-9+/]+={0,2}
```
