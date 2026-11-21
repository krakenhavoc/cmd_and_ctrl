package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// foretell_cards_test.go — #658, the three cards, against the REAL
// catalog. The model and the verb are pinned in the game package;
// what only a catalog test can check is that each card file actually
// declared foretell, and declared it at the printed cost.

const (
	SawItComingOracle        = "90edaf33-d0ab-47e0-8f6a-6fba38286e6e"
	BeholdTheMultiverseOracl = "d7d2f701-77df-4169-bf98-0d51d6886e9b"
)

// foretellOf reads the card's declared foretell action, or nil.
func ForetellOf(oracle string) *game.SpecialAction {
	for _, sa := range game.SpecialActionsFor(oracle) {
		if sa.Kind == game.SpecialActionForetell {
			out := sa
			return &out
		}
	}
	return nil
}

// containsFold is a case-insensitive substring test, kept local so the
// assertion above reads as one line.
func ContainsFold(haystack, needle string) bool {
	h, n := []rune(haystack), []rune(needle)
	lower := func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + 32
		}
		return r
	}
	if len(n) == 0 || len(n) > len(h) {
		return false
	}
	for i := 0; i+len(n) <= len(h); i++ {
		match := true
		for j := range n {
			if lower(h[i+j]) != lower(n[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
