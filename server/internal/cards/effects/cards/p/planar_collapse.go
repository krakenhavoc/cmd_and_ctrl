package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Planar Collapse — Enchantment {1}{W}:
//
//	"At the beginning of your upkeep, if there are four or more creatures
//	 on the battlefield, sacrifice this enchantment and destroy all
//	 creatures. They can't be regenerated."
//
// The count is every creature, whoever controls it (CR 603.4: checked as the
// ability triggers and as it resolves).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "82bccc9f-f700-4f1f-b89e-bea63d973161",
		Name:         "Planar Collapse",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.OldB717UpkeepWipe("Planar Collapse — sacrifice it and destroy all creatures", 4, effects.Creature(), true),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}}),
		},
	})
}
