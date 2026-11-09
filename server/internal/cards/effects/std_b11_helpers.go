package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b11_helpers.go — shared vocabulary of the Standard leftovers,
// batches 11 and 12 (#4452).

// sb11ColorBits gives each colour a bit of a SearchRule's number.
var sb11ColorBits = map[string]int{"W": 1, "U": 2, "B": 4, "R": 8, "G": 16}

// sb11ColorMask is the number of searchRuleOneCardPerColor naming these
// colours.
func sb11ColorMask(colors ...string) int {
	m := 0
	for _, c := range colors {
		m |= sb11ColorBits[c]
	}
	return m
}

// searchRuleOneCardPerColor is "search your library for a black card, a
// green card, and a blue card" (Lotuslight Dancers): the number names the
// colours, and the picked set must give every card its own colour, a card
// standing for one quality only. A card of several colours may be any one
// of them, and the others still need cards of their own (Lotuslight
// Dancers' ruling). Fewer cards than colours is fine, none at all too.
var searchRuleOneCardPerColor = game.RegisterSearchRule("search/one-card-per-color",
	func(picked []game.Card, mask int) bool {
		var colors []string
		for c, bit := range sb11ColorBits {
			if mask&bit != 0 {
				colors = append(colors, c)
			}
		}
		if len(picked) > len(colors) {
			return false
		}
		return sb11Assign(picked, colors, map[string]bool{})
	})

// sb11Assign reports whether every card in picked can take a distinct
// colour from colors.
func sb11Assign(picked []game.Card, colors []string, used map[string]bool) bool {
	if len(picked) == 0 {
		return true
	}
	for _, col := range colors {
		if used[col] || !picked[0].HasColor(col) {
			continue
		}
		used[col] = true
		ok := sb11Assign(picked[1:], colors, used)
		used[col] = false
		if ok {
			return true
		}
	}
	return false
}
