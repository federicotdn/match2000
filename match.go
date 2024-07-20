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
	"math"
	"math/big"
)

type Match struct {
	MatchThreshold float64
	MatchDistance  int
}

func NewMatch() *Match {
	return &Match{
		// How far to search for a match (0 = exact location, 1000+ = broad match). A
		// match this many characters away from the expected location will add 1.0 to
		// the score (0.0 is a perfect match).
		MatchDistance: 1000,
		// At what point is no match declared (0.0 = perfection, 1.0 = very loose).
		MatchThreshold: 0.5,
	}
}

func copyBigInt(z *big.Int) *big.Int {
	tmp := &big.Int{}
	tmp.Set(z)
	return tmp
}

// MatchMain locates the best instance of 'pattern' in 'text' near 'loc'.
// Returns -1 if no match found.
func (m *Match) MatchMain(text, pattern string, loc int) int {
	// Check for null inputs not needed since null can't be passed in Go.

	// Clamp loc between 0 and len(text)
	loc = max(0, min(loc, len(text)))
	switch {
	case text == pattern:
		// Shortcut (potentially not guaranteed by the algorithm)
		return 0
	case len(text) == 0:
		// Nothing to match.
		return -1
	case loc+len(pattern) <= len(text) && text[loc:loc+len(pattern)] == pattern:
		// Perfect match at the perfect spot!  (Includes case of null pattern)
		return loc
	}
	// Do a fuzzy compare.
	return m.MatchBitap(text, pattern, loc)
}

// MatchBitap locates the best instance of 'pattern' in 'text' near 'loc' using the
// Bitap algorithm.
// Returns -1 if no match was found.
func (m *Match) MatchBitap(text, pattern string, loc int) int {
	// Initialise the alphabet.
	s := m.matchAlphabet(pattern)

	// Highest score beyond which we give up.
	scoreThreshold := m.MatchThreshold
	// Is there a nearby exact match? (speedup)
	bestLoc := indexOf(text, pattern, loc)
	if bestLoc != -1 {
		scoreThreshold = min(m.matchBitapScore(0, bestLoc, loc, pattern), scoreThreshold)

		// What about in the other direction? (speedup)
		bestLoc = lastIndexOf(text, pattern, loc+len(pattern))
		if bestLoc != -1 {
			scoreThreshold = min(m.matchBitapScore(0, bestLoc, loc, pattern), scoreThreshold)
		}
	}

	// Initialise the bit arrays.
	matchMask := &big.Int{}
	matchMask.SetBit(matchMask, len(pattern)-1, 1)
	bestLoc = -1

	var binMin, binMid int
	binMax := len(pattern) + len(text)
	lastRd := []big.Int{}
	one := big.NewInt(1)

	for d := 0; d < len(pattern); d++ {
		// Scan for the best match; each iteration allows for one more error. Run a
		// binary search to determine how far from 'loc' we can stray at this error
		// level.
		binMin = 0
		binMid = binMax
		for binMin < binMid {
			if m.matchBitapScore(d, loc+binMid, loc, pattern) <= scoreThreshold {
				binMin = binMid
			} else {
				binMax = binMid
			}
			binMid = (binMax-binMin)/2 + binMin
		}
		// Use the result from this iteration as the maximum for the next.
		binMax = binMid
		start := max(1, loc-binMid+1)
		finish := min(loc+binMid, len(text)) + len(pattern)

		rd := make([]big.Int, finish+2)

		dShifted := &big.Int{}
		dShifted.SetBit(dShifted, d, 1).Sub(dShifted, one)
		rd[finish+1].Set(dShifted)

		for j := finish; j >= start; j-- {
			charMatch := &big.Int{}
			if j-1 < len(text) {
				val, ok := s[text[j-1]]
				if ok {
					charMatch.Set(val)
				}
			}

			// First pass: exact match.
			tmp := copyBigInt(&rd[j+1])
			tmp.Lsh(tmp, 1).Or(tmp, one).And(tmp, charMatch)
			rd[j].Set(tmp)

			if d > 0 {
				// Subsequent passes: fuzzy match.
				tmp = copyBigInt(&lastRd[j+1])
				tmp.Or(tmp, &lastRd[j]).Lsh(tmp, 1).Or(tmp, one).Or(tmp, &lastRd[j+1])
				rd[j].Or(&rd[j], tmp)
			}

			tmp = copyBigInt(&rd[j])
			tmp.And(tmp, matchMask)
			if tmp.Sign() != 0 {
				score := m.matchBitapScore(d, j-1, loc, pattern)
				// This match will almost certainly be better than any existing match.  But check anyway.
				if score <= scoreThreshold {
					// Told you so.
					scoreThreshold = score
					bestLoc = j - 1
					if bestLoc > loc {
						// When passing loc, don't exceed our current distance from loc.
						start = max(1, 2*loc-bestLoc)
					} else {
						// Already passed loc, downhill from here on in.
						break
					}
				}
			}
		}
		if m.matchBitapScore(d+1, loc, loc, pattern) > scoreThreshold {
			// No hope for a (better) match at greater error levels.
			break
		}
		lastRd = rd
	}
	return bestLoc
}

// matchBitapScore computes and returns the score for a match with e errors and x location.
// Lower scores are better.
func (m *Match) matchBitapScore(e, x, loc int, pattern string) float64 {
	accuracy := float64(e) / float64(len(pattern))
	proximity := math.Abs(float64(loc - x))
	if m.MatchDistance == 0 {
		// Dodge divide by zero error.
		if proximity == 0 {
			return accuracy
		}

		return 1.0
	}
	return accuracy + (proximity / float64(m.MatchDistance))
}

// matchAlphabet initialises the alphabet for the Bitap algorithm.
func (m *Match) matchAlphabet(pattern string) map[byte]*big.Int {
	s := map[byte]*big.Int{}
	charPattern := []byte(pattern)

	for i, c := range charPattern {
		val, ok := s[c]
		if !ok {
			val = &big.Int{}
			s[c] = val
		}
		val.SetBit(val, len(pattern)-i-1, 1)
	}
	return s
}
