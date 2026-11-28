package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elkin Bottle — Artifact {3} (Ice Age, S82 OLD #5636):
//
//	"{3}, {T}: Exile the top card of your library. Until the beginning of your
//	 next upkeep, you may play that card."
//
// Alania's Pathmaker's impulse window (ADR 0063), ending as the controller's next
// upkeep begins (CR 503.1a) rather than at end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e381361-7404-4654-bfb4-5ea65f6e4dce",
		Name:         "Elkin Bottle",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Exile the top card of your library. Until the beginning of your next upkeep, you may play that card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect:  effects.ElkinBottleActivation,
		}},
	})
}
