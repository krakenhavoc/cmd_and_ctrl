package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rushwood Elemental — Creature — Elemental {G}{G}{G}{G}{G}, 4/4 (Mercadian
// Masques, S82 OLD batch 698, #5619):
//
//	"Trample
//	 At the beginning of your upkeep, you may put a +1/+1 counter on this
//	 creature."
//
// Instrument of the Bards' optional upkeep counter: a real "may" asked as the
// trigger resolves, and nothing if the Elemental has left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d98edd47-c9de-46e9-a184-ba3bc0be42c4",
		Name:            "Rushwood Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Rushwood Elemental — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
				"Rushwood Elemental — put a +1/+1 counter on it?"),
		},
	})
}
