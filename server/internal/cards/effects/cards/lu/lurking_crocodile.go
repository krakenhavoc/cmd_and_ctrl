package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lurking Crocodile — Creature — Crocodile {2}{G}, 2/2 (Magic 2012, #5573):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)
//	 Islandwalk (This creature can't be blocked as long as defending
//	 player controls an Island.)"
//
// Bloodthirst is the shared entry replacement (bloodthirst.go); islandwalk
// is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b0df3389-5959-4ee2-a921-9324732f7ce8",
		Name:            "Lurking Crocodile",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Replacements:    []game.ReplacementEffect{effects.Bloodthirst(1, "Lurking Crocodile")},
	})
}
