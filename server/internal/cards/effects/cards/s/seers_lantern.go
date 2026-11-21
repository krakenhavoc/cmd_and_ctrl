package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "4b3dd7e5-e0b5-4773-bb16-37b757ae74d8",
		Name:          "Seer's Lantern",
		Completeness:  effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{effects.PainlessColorless()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
