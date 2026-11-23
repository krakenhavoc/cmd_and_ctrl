package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icecave Crasher — Creature — Beast {3}{G}, 4/4 (Edge of Eternities,
// #4461):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +1/+0 until end of turn."
//
// Attercop's landfall pump with +1/+0: one land, one trigger, each
// lasting until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b9650a0e-5795-42ff-84c6-6040e5cf4c90",
		Name:            "Icecave Crasher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Icecave Crasher — +1/+0 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(1, 0, "Icecave Crasher — +1/+0 until end of turn")),
		},
	})
}
