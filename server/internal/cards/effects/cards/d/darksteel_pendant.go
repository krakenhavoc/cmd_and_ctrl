package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darksteel Pendant — Artifact {2} (Darksteel, #5605):
//
//	"Indestructible (Effects that say "destroy" don't destroy this
//	 artifact.)
//	 {1}, {T}: Scry 1. (Look at the top card of your library. You may put
//	 that card on the bottom.)"
//
// Indestructible is a printed keyword (CR 702.12). The scry is Aether
// Theorist's Do(Scry{N: 1}) body behind Treasure Map's {1}, {T} cost; it
// answers nothing (AnswerValue, ADR 0142) and declares its scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "431838a8-f020-4e4e-a6f4-2d4ca27c56df",
		Name:            "Darksteel Pendant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Scry 1.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
