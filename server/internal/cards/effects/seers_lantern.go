package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seer's Lantern — Artifact {3} (Oath of the Gatewatch, #5545):
//
//	"{T}: Add {C}.
//	 {2}, {T}: Scry 1. (Look at the top card of your library. You may
//	 put that card on the bottom.)"
//
// The {C} ability is Eldrazi Temple's first row; the scry is Sonic
// Screwdriver's {2}, {T} row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "4b3dd7e5-e0b5-4773-bb16-37b757ae74d8",
		Name:          "Seer's Lantern",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Effect:  Do(Scry{N: 1}),
		}},
	})
}
