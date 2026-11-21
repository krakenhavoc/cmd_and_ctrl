package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Remnant Elemental — Creature — Elemental {1}{R}, 0/4 (Edge of
// Eternities, #4461):
//
//	"Reach
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +2/+0 until end of turn."
//
// Icecave Crasher's landfall pump with +2/+0: one land, one trigger,
// each lasting until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69e8333a-bddf-49e2-9c39-3dc46fec81e0",
		Name:            "Remnant Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Remnant Elemental — +2/+0 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 0, "Remnant Elemental — +2/+0 until end of turn")),
		},
	})
}
