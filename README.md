# regen

A Regular Expression GENerator. It aims to provide a fluent syntax for
constructing composable Go regular expressions in a way that that is easy to read 
and write at the cost of being much more verbose.

## Usage

```
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
    // Results in: ^(?P<greeting>[\p{Hiragana}\p{Katakana}\p{Han}]+|[a-zA-Z]+)(\s+(world|世界))?$

    re.MatchString("こんにちは")   // == true (`greeting` capture group == "こんにちは")
    re.MatchString("hello 世界")  // == true (`greeting` capture group == "hello")
}
```

If it is too awkward to construct your regular expression using this syntax,
but still want the other benefits of composition, use the `regen.Raw`
function to input raw regular expressions. For instance, to match base64 encoded
strings:

```
re := regexp.MustCompile(regen.Sequence(
    regen.Raw(`[A-Za-z0-9+/]`).Repeat().Min(1),
    regen.String("=").Repeat().Max(2),
))
// Results in: [A-Za-z0-9+/]+={0,2}
```

Note that this *could* be expressed without using `regen.Raw` as follows:

```
re := regexp.MustCompile(regen.Sequence(
    regen.CharRange('A', 'Z').Append(
        regen.CharRange('a', 'z'),
        regen.CharRange('0', '9'),
        regen.CharSet('+', '/'),
    ).Repeat().Min(1),
    regen.String("=").Repeat().Max(2),
))
// Results in: [A-Za-z0-9+/]+={0,2}
```

## Installation

```
$ go get github.com/aoldershaw/regen 
```
