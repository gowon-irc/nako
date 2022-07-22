package main

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/logrusorgru/aurora"
)

func containsString(ss []string, s string) bool {
	for _, i := range ss {
		if i == s {
			return true
		}
	}
	return false
}

func prefixValue(name string) int {
	prefixMap := map[byte]int{
		'~': 5,
		'&': 4,
		'@': 3,
		'%': 2,
		'+': 1,
	}

	v := prefixMap[name[0]]

	return v
}

func sortNamesList(names []string) []string {
	nc := make([]string, len(names))
	copy(nc, names)

	sort.Slice(nc, func(i, j int) bool {
		n1, n2 := nc[i], nc[j]
		pv1, pv2 := prefixValue(n1), prefixValue(n2)

		if pv1 == pv2 {
			return n1 < n2
		}

		return pv1 > pv2
	})

	return nc
}

func colourNamesList(names string, colourAllocator func(s string) uint8) string {
	namesList := sortNamesList(strings.Fields(names))

	colouredNames := []string{}

	for _, name := range namesList {
		sanitisedNick := strings.TrimLeftFunc(name, func(r rune) bool {
			return !unicode.IsLetter(r)
		})
		index := colourAllocator(sanitisedNick)
		colouredName := aurora.Index(index, name).String()
		colouredNames = append(colouredNames, colouredName)
	}

	return strings.Join(colouredNames, " ")
}

func genColourAllocator(seed int) func(s string) uint8 {
	m := make(map[string]uint8)

	return func(s string) uint8 {
		v, p := m[s]
		if p {
			return v
		}

		sum := 0
		for _, c := range s {
			sum += int(c)
		}

		rand.Seed(int64(seed + sum))
		id := uint8(rand.Intn(6) + 1)
		m[s] = id
		return id
	}
}

func ircToAnsiColours(s string) string {
	out := s

	m := map[string]string{
		"\u000301": "\x1b[30m", // red
		"\u000302": "\x1b[34m", // navy blue -> blue
		"\u000303": "\x1b[32m", // green
		"\u000304": "\x1b[31m", // red
		"\u000305": "\x1b[33m", // brown -> yellow
		"\u000306": "\x1b[35m", // purple -> magenta
		"\u000307": "\x1b[32m", // olive -> green
		"\u000308": "\x1b[33m", // yellow
		"\u000309": "\x1b[32m", // lime green -> green
		"\u000310": "\x1b[36m", // teal -> cyan
		"\u000311": "\x1b[36m", // aqua blue -> cyan
		"\u000312": "\x1b[34m", // royal blue -> blue
		"\u000313": "\x1b[31m", // hot pink -> red
		"\u000314": "\x1b[30m", // dark grey -> black
		"\u000315": "\x1b[30m", // light grey -> black
		"\u000316": "\x1b[37m", // white
		"\u000399": "\x1b[0m",  // reset
	}

	for k, v := range m {
		out = strings.Replace(out, k, v, -1)
	}

	return out
}

func getCommand(s string) (command string, args []string) {
	if !strings.HasPrefix(s, "/") {
		return "", []string{}
	}

	fields := strings.Fields(s)
	command = strings.TrimPrefix(fields[0], "/")

	return command, fields[1:]
}

func stringIsNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
