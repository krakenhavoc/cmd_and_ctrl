package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thran War Machine — Artifact Creature — Construct, {4}, 4/5:
//
//	"Echo {4} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 This creature attacks each combat if able."
//
// "Attacks each combat if able" is AttacksEachCombat() (CR 508.1d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b3b8fe34-4f69-4555-bcac-04563d02c99f",
		Name:         "Thran War Machine",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.AttacksEachCombat()},
		Triggered: []game.TriggeredAbility{
			effects.Echo("Thran War Machine", "{4}"),
		},
	})
}
