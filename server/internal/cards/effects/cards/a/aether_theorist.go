package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Theorist — Creature — Vedalken Rogue {1}{U}, 1/3:
//
//	"When this creature enters, you get {E}{E}{E} (three energy counters).
//	 {T}, Pay {E}: Scry 1. (Look at the top card of your library. You may
//	 put that card on the bottom.)"
//
// ADR 0129 PR 1 (#1995): the energy is a counter on the controller
// (CR 107.14), paid as an activation cost component.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b07ef53e-46db-457e-a5a2-846bb9805b70",
		Name:         "Aether Theorist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Aether Theorist", 3),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay {E}: Scry 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Scry: 1},
			Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(1)),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
