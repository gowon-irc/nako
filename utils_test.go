package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainsString(t *testing.T) {
	cases := []struct {
		name     string
		ss       []string
		s        string
		expected bool
	}{
		{
			name:     "string in slice",
			ss:       []string{"abc", "def", "ghi"},
			s:        "abc",
			expected: true,
		},
		{
			name:     "string not in slice",
			ss:       []string{"abc", "def", "ghi"},
			s:        "jkl",
			expected: false,
		},
		{
			name:     "empty slice",
			ss:       []string{},
			s:        "abc",
			expected: false,
		},
		{
			name:     "empty string arg",
			ss:       []string{"abc", "def", "ghi"},
			s:        "",
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := containsString(tc.ss, tc.s)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestPrefixValue(t *testing.T) {
	cases := []struct {
		name     string
		n        string
		expected int
	}{
		{
			name:     "prefix in map",
			n:        "~nako",
			expected: 5,
		},
		{
			name:     "prefix not in map",
			n:        "*nako",
			expected: 0,
		},
		{
			name:     "no prefix",
			n:        "nako",
			expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := prefixValue(tc.n)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestSortNamesList(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		out  []string
	}{
		{
			name: "empty list",
			in:   []string{},
			out:  []string{},
		},
		{
			name: "already sorted list",
			in:   []string{"~nako", "&a", "b", "c"},
			out:  []string{"~nako", "&a", "b", "c"},
		},
		{
			name: "no ops",
			in:   []string{"b", "a", "c"},
			out:  []string{"a", "b", "c"},
		},
		{
			name: "one op",
			in:   []string{"b", "@nako", "c"},
			out:  []string{"@nako", "b", "c"},
		},
		{
			name: "two ops",
			in:   []string{"@nako", "@a", "c"},
			out:  []string{"@a", "@nako", "c"},
		},
		{
			name: "one op, one hop",
			in:   []string{"b", "%c", "@nako"},
			out:  []string{"@nako", "%c", "b"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sortNamesList(tc.in)
			assert.Equal(t, tc.out, got)
		})
	}
}

func TestColourNamesLi(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "empty string",
			in:   "",
			out:  "",
		},
		{
			name: "one name",
			in:   "nako",
			out:  "\x1b[33mnako\x1b[0m",
		},
		{
			name: "one prefixed name",
			in:   "~nako",
			out:  "\x1b[33m~nako\x1b[0m",
		},
		{
			name: "one prefixed, one normal",
			in:   "~nako a",
			out:  "\x1b[33m~nako\x1b[0m \x1b[31ma\x1b[0m",
		},
	}

	colourAllocator := genColourAllocator(1)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := colourNamesList(tc.in, colourAllocator)
			assert.Equal(t, tc.out, got)
		})
	}
}

func TestGenColourAllocator(t *testing.T) {
	cases := []struct {
		name string
		in   int
		out  uint8
	}{
		{
			name: "1 as seed",
			in:   1,
			out:  3,
		},
		{
			name: "2 as seed",
			in:   2,
			out:  6,
		},
		{
			name: "-1 as seed",
			in:   -1,
			out:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			colourAllocator := genColourAllocator(tc.in)
			got := colourAllocator("nako")
			assert.Equal(t, tc.out, got)
		})
	}
}

func TestIrcToAnsiColours(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  string
	}{
		{
			name: "empty string",
			in:   "",
			out:  "",
		},
		{
			name: "no colours",
			in:   "nako",
			out:  "nako",
		},
		{
			name: "one colour",
			in:   "\u000301nako",
			out:  "\x1b[30mnako",
		},
		{
			name: "colour and reset",
			in:   "\u000301nako\u000399",
			out:  "\x1b[30mnako\x1b[0m",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ircToAnsiColours(tc.in)
			assert.Equal(t, tc.out, got)
		})
	}
}

func TestGetCommand(t *testing.T) {
	cases := []struct {
		name    string
		in      string
		command string
		args    []string
	}{
		{
			name:    "empty string",
			in:      "",
			command: "",
			args:    []string{},
		},
		{
			name:    "no command",
			in:      "nako",
			command: "",
			args:    []string{},
		},
		{
			name:    "command, no args",
			in:      "/nako",
			command: "nako",
			args:    []string{},
		},
		{
			name:    "command and args",
			in:      "/nako 3",
			command: "nako",
			args:    []string{"3"},
		},
	}

	for _, tc := range cases {
		command, args := getCommand(tc.in)
		assert.Equal(t, tc.command, command)
		assert.Equal(t, tc.args, args)
	}
}

func TestStringIsNumber(t *testing.T) {
	cases := []struct {
		name string
		in   string
		out  bool
	}{
		{
			name: "number input",
			in:   "123",
			out:  true,
		},
		{
			name: "no number input",
			in:   "nako",
			out:  false,
		},
		{
			name: "numbers and text",
			in:   "123nako123",
			out:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := stringIsNumber(tc.in)
			assert.Equal(t, tc.out, out)
		})
	}
}
