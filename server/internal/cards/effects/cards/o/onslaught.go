package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Onslaught — Enchantment {R} (Exodus, S82 OLD batch 743, #5626):
//
//	"Whenever you cast a creature spell, tap target creature."
//
// Beast Whisperer's cast trigger with a target: your own creature spells only,
// the target chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b); it taps any creature, yours included.
//
// No simplification.
func init() {
	const label = "Onslaught — tap target creature"
	effects.Register(effects.Spec{
		OracleID:     "ae9ca82c-e07e-4a41-a387-0ef7d6df14b6",
		Name:         "Onslaught",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, effects.YouCastACreatureSpell, label, effects.ThousandMoonsCrackshotTap),
				effects.TargetCreature("target creature")),
		},
	})
}
