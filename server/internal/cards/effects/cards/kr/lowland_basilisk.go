package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lowland Basilisk — Creature — Basilisk {2}{G}, 1/3 (Stronghold, S82 OLD batch
// 747, #5627):
//
//	"Whenever this creature deals damage to a creature, destroy that creature
//	 at end of combat."
//
// Ohran Viper's shape on any damage rather than combat damage alone: the
// damaged creature is recorded as the object it was (CR 400.7) as the trigger
// goes on the stack, and a delayed trigger at the beginning of the end of
// combat step (CR 511.2) destroys that object if it is still on the
// battlefield. Damage outside combat waits for the next end of combat step to
// begin (CR 603.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "effe239e-ade0-4ee2-99d7-f93b2f5ef585",
		Name:         "Lowland Basilisk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.DestroyTheDamagedCreatureAtEndOfCombat("Lowland Basilisk — destroy that creature at end of combat"),
		},
	})
}
