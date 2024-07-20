// Copyright (c) 2012-2016 The go-diff authors. All rights reserved.
// https://github.com/sergi/go-diff
// See the included LICENSE file for license details.
//
// go-diff is a Go implementation of Google's Diff, Match, and Patch library
// Original library is Copyright (c) 2006 Google Inc.
// http://code.google.com/p/google-diff-match-patch/

// Modified for match2000.

package match2000

import (
	"maps"
	"math/big"
	"testing"
)

// 1226 characters (ASCII)
const longText = `I saw Flying Lotus at a grocery store in Los Angeles yesterday. I told him how cool it was to meet him in person, but I didn't want to be a douche and bother him and ask him for photos or anything.
He said, "Oh, like you're doing now?"
I was taken aback, and all I could say was "Huh?" but he kept cutting me off and going "huh? huh? huh?" and closing his hand shut in front of my face. I walked away and continued with my shopping, and I heard him chuckle as I walked off. When I came to pay for my stuff up front I saw him trying to walk out the doors with like fifteen Milky Ways in his hands without paying.
The girl at the counter was very nice about it and professional, and was like "Sir, you need to pay for those first." At first he kept pretending to be tired and not hear her, but eventually turned back around and brought them to the counter.
When she took one of the bars and started scanning it multiple times, he stopped her and told her to scan them each individually "to prevent any electrical infetterence" and then turned around and winked at me. I don't even think that's a word. After she scanned each bar and put them in a bag and started to say the price, he kept interrupting her by yawning really loudly.`

// 211 characters (ASCII)
const longPattern = `When she took one of the bars and started scanning it multiple times, he stopped her and told her to scan them each individually "to prevent any electrical infetterence" and then turned around and winked at me.`

// Same as above, slightly modified
// 210 characters (ASCII)
const longPattern2 = `When she took one of the bars and started scanning it multiple times he stopped her and told her to scan them each individually "to prevent all electrical interference" and then turned around and winked at me.`

func AlphabetsEqual(a, b map[byte]*big.Int) bool {
	return maps.EqualFunc(a, b, func(x, y *big.Int) bool {
		return x.Cmp(y) == 0
	})
}

func TestMatchAlphabet(t *testing.T) {
	type TestCase struct {
		Pattern string

		Expected map[byte]*big.Int
	}

	dmp := NewMatch()

	for i, tc := range []TestCase{
		{
			Pattern: "abc",

			Expected: map[byte]*big.Int{
				'a': big.NewInt(4),
				'b': big.NewInt(2),
				'c': big.NewInt(1),
			},
		},
		{
			Pattern: "abcaba",

			Expected: map[byte]*big.Int{
				'a': big.NewInt(37),
				'b': big.NewInt(18),
				'c': big.NewInt(8),
			},
		},
	} {
		actual := dmp.matchAlphabet(tc.Pattern)
		if !AlphabetsEqual(tc.Expected, actual) {
			t.Errorf("Test case #%d, %#v", i, tc)
		}
	}
}

func TestMatchBitap(t *testing.T) {
	type TestCase struct {
		Name string

		Text     string
		Pattern  string
		Location int

		Expected int
	}

	dmp := NewMatch()
	dmp.MatchDistance = 100
	dmp.MatchThreshold = 0.5

	for i, tc := range []TestCase{
		{"Exact match #1", "abcdefghijk", "fgh", 5, 5},
		{"Exact match #2", "abcdefghijk", "fgh", 0, 5},
		{"Fuzzy match #1", "abcdefghijk", "efxhi", 0, 4},
		{"Fuzzy match #2", "abcdefghijk", "cdefxyhijk", 5, 2},
		{"Fuzzy match #3", "abcdefghijk", "bxy", 1, -1},
		{"Overflow", "123456789xx0", "3456789x0", 2, 2},
		{"Before start match", "abcdef", "xxabc", 4, 0},
		{"Beyond end match", "abcdef", "defyy", 4, 3},
		{"Oversized pattern", "abcdef", "xabcdefy", 0, 0},
	} {
		actual := dmp.MatchBitap(tc.Text, tc.Pattern, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s", i, tc.Name)
		}
	}

	dmp.MatchThreshold = 0.4

	for i, tc := range []TestCase{
		{"Threshold #1", "abcdefghijk", "efxyhi", 1, 4},
	} {
		actual := dmp.MatchBitap(tc.Text, tc.Pattern, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s", i, tc.Name)
		}
	}

	dmp.MatchThreshold = 0.3

	for i, tc := range []TestCase{
		{"Threshold #2", "abcdefghijk", "efxyhi", 1, -1},
	} {
		actual := dmp.MatchBitap(tc.Text, tc.Pattern, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s", i, tc.Name)
		}
	}

	dmp.MatchThreshold = 0.0

	for i, tc := range []TestCase{
		{"Threshold #3", "abcdefghijk", "bcdef", 1, 1},
	} {
		actual := dmp.MatchBitap(tc.Text, tc.Pattern, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s", i, tc.Name)
		}
	}

	dmp.MatchThreshold = 0.5

	for i, tc := range []TestCase{
		{"Multiple select #1", "abcdexyzabcde", "abccde", 3, 0},
		{"Multiple select #2", "abcdexyzabcde", "abccde", 5, 8},
	} {
		actual := dmp.MatchBitap(tc.Text, tc.Pattern, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s", i, tc.Name)
		}
	}

	// Strict location.
	dmp.MatchDistance = 10

	for i, tc := range []TestCase{
		{"Distance test #1", "abcdefghijklmnopqrstuvwxyz", "abcdefg", 24, -1},
		{"Distance test #2", "abcdefghijklmnopqrstuvwxyz", "abcdxxefg", 1, 0},
	} {
		actual := dmp.MatchBitap(tc.Text, tc.Pattern, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s", i, tc.Name)
		}
	}

	// Loose location.
	dmp.MatchDistance = 1000

	for i, tc := range []TestCase{
		{"Distance test #3", "abcdefghijklmnopqrstuvwxyz", "abcdefg", 24, 0},
	} {
		actual := dmp.MatchBitap(tc.Text, tc.Pattern, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s", i, tc.Name)
		}
	}

	// Very long pattern
	dmp.MatchDistance = 1000
	dmp.MatchThreshold = 0.1

	for i, tc := range []TestCase{
		{"Long pattern test #1", longText, longPattern, 850, 855},
		{"Long pattern test #2", longText, longPattern2, 851, 855},
	} {
		actual := dmp.MatchBitap(tc.Text, tc.Pattern, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s, actual %v", i, tc.Name, actual)
		}
	}
}

func TestMatchMain(t *testing.T) {
	type TestCase struct {
		Name string

		Text1    string
		Text2    string
		Location int

		Expected int
	}

	dmp := NewMatch()

	for i, tc := range []TestCase{
		{"Equality", "abcdef", "abcdef", 1000, 0},
		{"Null text", "", "abcdef", 1, -1},
		{"Null pattern", "abcdef", "", 3, 3},
		{"Exact match", "abcdef", "de", 3, 3},
		{"Beyond end match", "abcdef", "defy", 4, 3},
		{"Oversized pattern", "abcdef", "abcdefy", 0, 0},
	} {
		actual := dmp.MatchMain(tc.Text1, tc.Text2, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s", i, tc.Name)
		}
	}

	dmp.MatchThreshold = 0.7

	for i, tc := range []TestCase{
		{"Complex match", "I am the very model of a modern major general.", " that berry ", 5, 4},
	} {
		actual := dmp.MatchMain(tc.Text1, tc.Text2, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %#v", i, tc)
		}
	}

	// Very long pattern
	dmp.MatchDistance = 0
	dmp.MatchThreshold = 0.0

	for i, tc := range []TestCase{
		{"Long pattern test #1", longText, longPattern, 855, 855},
	} {
		actual := dmp.MatchMain(tc.Text1, tc.Text2, tc.Location)
		if tc.Expected != actual {
			t.Errorf("Test case #%d, %s, actual %v", i, tc.Name, actual)
		}
	}
}
