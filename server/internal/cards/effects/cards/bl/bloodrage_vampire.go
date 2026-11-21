package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodrage Vampire — Creature — Vampire {2}{B}, 3/1 (Magic 2012, #5573):
//
//	"Bloodthirst 1 (If an opponent was dealt damage this turn, this
//	 creature enters with a +1/+1 counter on it.)"
//
// Bloodthirst is the shared entry replacement (bloodthirst.go), the entry
// replacement and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1b6a840-2dc6-4af3-869a-b3bfe64ff7c0",
		Name:         "Bloodrage Vampire",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.Bloodthirst(1, "Bloodrage Vampire")},
	})
}
