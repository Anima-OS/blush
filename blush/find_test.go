package blush_test

import (
	"regexp"
	"testing"

	"github.com/arsham/blush/blush"
)

type colourer interface {
	Colour() blush.Colour
}

func TestNewLocatorColours(t *testing.T) {
	tcs := []struct {
		name   string
		colour string
		want   blush.Colour
	}{
		{"default", "", blush.DefaultColour},
		{"garbage", "sdsds", blush.DefaultColour},
		{"blue", "blue", blush.FgBlue},
		{"blue short", "b", blush.FgBlue},
		{"red", "red", blush.FgRed},
		{"red short", "r", blush.FgRed},
		{"blue", "blue", blush.FgBlue},
		{"blue short", "b", blush.FgBlue},
		{"green", "green", blush.FgGreen},
		{"green short", "g", blush.FgGreen},
		{"black", "black", blush.FgBlack},
		{"black short", "bl", blush.FgBlack},
		{"white", "white", blush.FgWhite},
		{"white short", "w", blush.FgWhite},
		{"cyan", "cyan", blush.FgCyan},
		{"cyan short", "cy", blush.FgCyan},
		{"magenta", "magenta", blush.FgMagenta},
		{"magenta short", "mg", blush.FgMagenta},
		{"yellow", "yellow", blush.FgYellow},
		{"yellow short", "yl", blush.FgYellow},
		{"no colour", "no-colour", blush.NoColour},
		{"no colour american", "no-color", blush.NoColour},

		{"hash 000", "#000", blush.Colour{R: 0, G: 0, B: 0}},
		{"hash 666", "#666", blush.Colour{R: 102, G: 102, B: 102}},
		{"hash 000000", "#000000", blush.Colour{R: 0, G: 0, B: 0}},
		{"hash 666666", "#666666", blush.Colour{R: 102, G: 102, B: 102}},
		{"hash FFF", "#FFF", blush.Colour{R: 255, G: 255, B: 255}},
		{"hash fff", "#fff", blush.Colour{R: 255, G: 255, B: 255}},
		{"hash ffffff", "#ffffff", blush.Colour{R: 255, G: 255, B: 255}},
		{"hash ababAB", "#ababAB", blush.Colour{R: 171, G: 171, B: 171}},
		{"hash hhhhhh", "#hhhhhh", blush.DefaultColour},
		{"hash aaaaaaa", "#aaaaaaa", blush.DefaultColour},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			l := blush.NewLocator(tc.colour, "aaa", false)
			c, ok := l.(colourer)
			if !ok {
				t.Fatalf("%v does not implement Colour() method", l)
			}
			if c.Colour() != tc.want {
				t.Errorf("%s: c.Colour() = %#v, want %#v", tc.colour, c.Colour(), tc.want)
			}
		})
	}
}

func TestNewLocatorExact(t *testing.T) {
	l := blush.NewLocator("", "aaa", false)
	if _, ok := l.(blush.Exact); !ok {
		t.Errorf("l = %T, want *blush.Exact", l)
	}
	l = blush.NewLocator("", "*aaa", false)
	if _, ok := l.(blush.Exact); !ok {
		t.Errorf("l = %T, want *blush.Exact", l)
	}
}

func TestNewLocatorIexact(t *testing.T) {
	l := blush.NewLocator("", "aaa", true)
	if _, ok := l.(blush.Iexact); !ok {
		t.Errorf("l = %T, want *blush.Iexact", l)
	}
	l = blush.NewLocator("", "*aaa", true)
	if _, ok := l.(blush.Iexact); !ok {
		t.Errorf("l = %T, want *blush.Iexact", l)
	}
}

func TestNewLocatorRx(t *testing.T) {
	tcs := []struct {
		name    string
		input   string
		matches []string
	}{
		{"empty", "^$", []string{""}},
		{"starts with", "^aaa", []string{"aaa", "aaa sss"}},
		{"ends with", "aaa$", []string{"aaa", "sss aaa"}},
		{"with star", "blah blah.*", []string{"blah blah", "aa blah blah aa"}},
		{"with curly brackets", "a{3}", []string{"aaa", "aa aaa aa"}},
		{"with brackets", "[ab]", []string{"kjhadf", "kjlrbrlkj", "sdbsdha"}},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			l := blush.NewLocator("", tc.input, false)
			if _, ok := l.(blush.Rx); !ok {
				t.Errorf("l = %T, want *blush.Rx", l)
			}
			l = blush.NewLocator("", tc.input, false)
			if _, ok := l.(blush.Rx); !ok {
				t.Errorf("l = %T, want *blush.Rx", l)
			}
		})
	}
}

func TestNewLocatorRxColours(t *testing.T) {
	rx := blush.NewLocator("b", "a{3}", false)
	want := "this " + blush.Colourise("aaa", blush.FgBlue) + "meeting"
	got, ok := rx.Find("this aaameeting")
	if got != want {
		t.Errorf("got = `%s`, want `%s`", got, want)
	}
	if !ok {
		t.Error("ok = false, want true")
	}
}

func TestExactNotFound(t *testing.T) {
	l := blush.NewExact("nooooo", blush.NoColour)
	got, ok := l.Find("yessss")
	if got != "" {
		t.Errorf("got = %s, want `%s`", got, "")
	}
	if ok {
		t.Error("ok = true, want false")
	}
}

func TestExactFind(t *testing.T) {
	l := blush.NewExact("nooooo", blush.NoColour)
	got, ok := l.Find("yessss")
	if got != "" {
		t.Errorf("got = %s, want `%s`", got, "")
	}
	if ok {
		t.Error("ok = true, want false")
	}

	tcs := []struct {
		name   string
		search string
		colour blush.Colour
		input  string
		want   string
		wantOk bool
	}{
		{"exact no colour", "aaa", blush.NoColour, "aaa", "aaa", true},
		{"exact not found", "aaaa", blush.NoColour, "aaa", "", false},
		{"some parts no colour", "aaa", blush.NoColour, "bb aaa bb", "bb aaa bb", true},
		{"exact blue", "aaa", blush.FgBlue, "aaa", blush.Colourise("aaa", blush.FgBlue), true},
		{"some parts blue", "aaa", blush.FgBlue, "bb aaa bb", "bb " + blush.Colourise("aaa", blush.FgBlue) + " bb", true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			l := blush.NewExact(tc.search, tc.colour)
			got, ok := l.Find(tc.input)
			if got != tc.want {
				t.Errorf("got = `%s`, want `%s`", got, tc.want)
			}
			if ok != tc.wantOk {
				t.Errorf("ok = %t, want %t", ok, tc.wantOk)
			}
		})
	}
}

func TestRxNotFound(t *testing.T) {
	l := blush.NewRx(regexp.MustCompile("nooooo"), blush.NoColour)
	got, ok := l.Find("yessss")
	if got != "" {
		t.Errorf("got = %s, want `%s`", got, "")
	}
	if ok {
		t.Error("ok = true, want false")
	}
}

func TestRxFind(t *testing.T) {
	l := blush.NewRx(regexp.MustCompile("nooooo"), blush.NoColour)
	got, ok := l.Find("yessss")
	if got != "" {
		t.Errorf("got = %s, want `%s`", got, "")
	}
	if ok {
		t.Error("ok = true, want false")
	}

	tcs := []struct {
		name   string
		search string
		colour blush.Colour
		input  string
		want   string
		wantOk bool
	}{
		{"exact no colour", "(^aaa$)", blush.NoColour, "aaa", "aaa", true},
		{"exact not found", "(^aa$)", blush.NoColour, "aaa", "", false},
		{"some parts no colour", "(aaa)", blush.NoColour, "bb aaa bb", "bb aaa bb", true},
		{"some parts not matched", "(Aaa)", blush.NoColour, "bb aaa bb", "", false},
		{"exact blue", "(aaa)", blush.FgBlue, "aaa", blush.Colourise("aaa", blush.FgBlue), true},
		{"some parts blue", "(aaa)", blush.FgBlue, "bb aaa bb", "bb " + blush.Colourise("aaa", blush.FgBlue) + " bb", true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			l := blush.NewRx(regexp.MustCompile(tc.search), tc.colour)
			got, ok := l.Find(tc.input)
			if got != tc.want {
				t.Errorf("got = `%s`, want `%s`", got, tc.want)
			}
			if ok != tc.wantOk {
				t.Errorf("ok = %t, want %t", ok, tc.wantOk)
			}
			if l.Colour() != tc.colour {
				t.Errorf("l.Colour() = %v, want %v", l.Colour(), tc.colour)
			}
		})
	}
}

func TestIexactNotFound(t *testing.T) {
	l := blush.NewIexact("nooooo", blush.NoColour)
	got, ok := l.Find("yessss")
	if got != "" {
		t.Errorf("got = %s, want `%s`", got, "")
	}
	if ok {
		t.Error("ok = true, want false")
	}
}

func TestIexact(t *testing.T) {
	tcs := []struct {
		name   string
		search string
		colour blush.Colour
		input  string
		want   string
		wantOk bool
	}{
		{"exact no colour", "aaa", blush.NoColour, "aaa", "aaa", true},
		{"exact not found", "aaaa", blush.NoColour, "aaa", "", false},
		{"i exact no colour", "AAA", blush.NoColour, "aaa", "aaa", true},
		{"some parts no colour", "aaa", blush.NoColour, "bb aaa bb", "bb aaa bb", true},
		{"i some parts no colour", "AAA", blush.NoColour, "bb aaa bb", "bb aaa bb", true},
		{"exact blue", "aaa", blush.FgBlue, "aaa", blush.Colourise("aaa", blush.FgBlue), true},
		{"i exact blue", "AAA", blush.FgBlue, "aaa", blush.Colourise("aaa", blush.FgBlue), true},
		{"some parts blue", "aaa", blush.FgBlue, "bb aaa bb", "bb " + blush.Colourise("aaa", blush.FgBlue) + " bb", true},
		{"i some parts blue", "AAA", blush.FgBlue, "bb aaa bb", "bb " + blush.Colourise("aaa", blush.FgBlue) + " bb", true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			l := blush.NewIexact(tc.search, tc.colour)
			got, ok := l.Find(tc.input)
			if got != tc.want {
				t.Errorf("got = `%s`, want `%s`", got, tc.want)
			}
			if ok != tc.wantOk {
				t.Errorf("ok = %t, want %t", ok, tc.wantOk)
			}
			if l.Colour() != tc.colour {
				t.Errorf("l.Colour() = %v, want %v", l.Colour(), tc.colour)
			}
		})
	}
}

func TestRxInsensitiveFind(t *testing.T) {
	tcs := []struct {
		name   string
		search string
		colour string
		input  string
		want   string
		wantOk bool
	}{
		{"exact no colour", "^AAA$", "no-colour", "aaa", "aaa", true},
		{"exact not found", "^AA$", "no-colour", "aaa", "", false},
		{"some words no colour", `AAA*`, "no-colour", "bb aaa bb", "bb aaa bb", true},
		{"exact blue", "^AAA$", "b", "aaa", blush.Colourise("aaa", blush.FgBlue), true},
		{"default colour", "^AAA$", "", "aaa", blush.Colourise("aaa", blush.DefaultColour), true},
		{"some words blue", "AAA?", "b", "bb aaa bb", "bb " + blush.Colourise("aaa", blush.FgBlue) + " bb", true},
		{"some words blue long", "AAA?", "blue", "bb aaa bb", "bb " + blush.Colourise("aaa", blush.FgBlue) + " bb", true},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			l := blush.NewLocator(tc.colour, tc.search, true)
			if _, ok := l.(blush.Rx); !ok {
				t.Fatalf("l = %T, want blush.Rx", l)
			}
			got, ok := l.Find(tc.input)
			if got != tc.want {
				t.Errorf("got = `%s`, want `%s`", got, tc.want)
			}
			if ok != tc.wantOk {
				t.Errorf("ok = %t, want %t", ok, tc.wantOk)
			}
		})
	}

	rx := blush.NewLocator("b", "A{3}", true)
	want := "this " + blush.Colourise("aaa", blush.FgBlue) + "meeting"
	got, ok := rx.Find("this aaameeting")
	if got != want {
		t.Errorf("got = `%s`, want `%s`", got, want)
	}
	if !ok {
		t.Error("ok = false, want true")
	}
}
