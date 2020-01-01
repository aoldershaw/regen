# regen

A Regular Expression GENerator. It aims to provide a fluent syntax for constructing composable
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
    japaneseWord := regen.UnicodeCharClass("Hiragana").Append(
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
    // Results in: ^(?P<greeting>[\p{Hiragana}\p{Katakana}\p{Han}]+|w+)(\s+(world|世界))?$

    re.MatchString("こんにちは")   // == true (`greeting` capture group == "こんにちは")
    re.MatchString("hello 世界")  // == true (`greeting` capture group == "hello")
}
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
...
```

Note that this *could* be expressed without using `regen.Raw` as follows:

```go
re := regexp.MustCompile(regen.Sequence(
    regen.CharRange('A', 'Z').Append(
        regen.CharRange('a', 'z'),
        regen.CharRange('0', '9'),
        regen.CharSet('+', '/'),
    ).Repeat().Min(1),
    regen.String("=").Repeat().Max(2),
).Regexp())
// Results in: [A-Za-z0-9+/]+={0,2}
...
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

Or you can unset flags if they were were set in a top-level group:

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
in the resulting regexp. For instance, the result `regen.OneOf` will be grouped. If you need to rely
on a particular ordering of capture groups, you should explicitly call `.Group().NoCapture()` on
sub expressions that should not be captured.
