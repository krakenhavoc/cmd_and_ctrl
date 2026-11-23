package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildfire Emissary — Creature — Efreet {3}{R}, 2/4 (Mirage, #5633):
//
//	"Protection from white
//	 {1}{R}: This creature gets +1/+0 until end of turn."
//
// Protection from white is a printed keyword token the engine honours
// for targeting, blocking, damage and attachment (CR 702.16). The pump
// is the shared thisGetsUntilEndOfTurn body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8fb44b03-017f-4cf3-a79b-8f6cc33570ca",
		Name:            "Wildfire Emissary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:    effects.ManaCost("{1}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Wildfire Emissary — +1/+0 until end of turn"),
		}},
	})
}
