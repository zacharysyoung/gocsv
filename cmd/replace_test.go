package cmd

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

func TestRunReplace(t *testing.T) {
	testCases := []struct {
		columnsString   string
		regex           string
		repl            string
		caseInsensitive bool
		rows            [][]string
	}{
		{"String", "Two", "Dos", false, [][]string{
			{"Number", "String"},
			{"1", "One"},
			{"2", "Dos"},
			{"-1", "Minus One"},
			{"2", "Another Dos"},
		}},
		{"String", "^one", "UNO", true, [][]string{
			{"Number", "String"},
			{"1", "UNO"},
			{"2", "Two"},
			{"-1", "Minus One"},
			{"2", "Another Two"},
		}},
	}
	for i, tt := range testCases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			ic, err := NewInputCsv("../test-files/simple-sort.csv")
			if err != nil {
				t.Error("Unexpected error", err)
			}
			toc := new(testOutputCsv)
			sub := new(ReplaceSubcommand)
			sub.columnsString = tt.columnsString
			sub.regex = tt.regex
			sub.repl = tt.repl
			sub.caseInsensitive = tt.caseInsensitive
			sub.RunReplace(ic, toc)
			err = assertRowsEqual(tt.rows, toc.rows)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestIssue78_unquote(t *testing.T) {
	testCases := []struct {
		regex, repl string
		elem        string

		want string
		err  error
	}{
		// literal
		{"z", `Z`, `baz`, `baZ`, nil},

		// codepoint escape sequences
		{"z", `\x5A`, `baz`, `baZ`, nil},
		{"z", `\u005A`, `baz`, `baZ`, nil},
		{"z", `\U0000005A`, `baz`, `baZ`, nil},

		// escaped escape chars
		{`"z"`, `\"Z\"`, `ba"z"`, `ba"Z"`, nil},
		{`"z"`, `\"\u005A\"`, `ba"z"`, `ba"Z"`, nil},
		{`z`, `\\u005A`, `baz`, `ba\u005A`, nil},
		{`"z"`, `\"\\u005A\"`, `ba"z"`, `ba"\u005A"`, nil},

		// with submatch names
		{`(foo).+`, `${1}baZ`, `foobar`, `foobaZ`, nil},
		{`(foo).+`, `${1}ba\u005A`, `foobar`, `foobaZ`, nil},
		{`(?P<FooName>foo).+`, `${FooName}ba\u005A`, `foobar`, `foobaZ`, nil},

		// bad input
		{`"z"`, `"Z"`, `ba"z"`, ``, strconv.ErrSyntax},
	}

	for _, tc := range testCases {
		f, err := regexReplaceFunc(tc.regex, tc.repl)
		if !errors.Is(err, tc.err) {
			t.Errorf("regexReplaceFunc(%s, %s)=%v; want %v",
				tc.regex, tc.repl, err, tc.err)
			continue
		}
		if tc.err != nil {
			continue
		}
		if got := f(tc.elem); got != tc.want {
			t.Errorf("regexReplaceFunc(%s, %s)(%s)=%s; want %s",
				tc.regex, tc.repl, tc.elem, got, tc.want)
		}
	}
}
