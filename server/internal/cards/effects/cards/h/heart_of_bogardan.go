package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heart of Bogardan — {2}{R}{R} Enchantment (Weatherlight, S82 OLD batch 772,
// #5629):
//
//	"Cumulative upkeep {2}
//	 When a player doesn't pay this enchantment's cumulative upkeep, this
//	 enchantment deals X damage to target player or planeswalker and each creature
//	 that player or that planeswalker's controller controls, where X is twice the
//	 number of age counters on this enchantment minus 2."
//
// The cumulative upkeep announces a payment it was declined (#6188,
// EventCumulativeUpkeepNotPaid) before the enchantment is sacrificed, so the
// trigger goes on the stack with its age counters on the event. The target is
// chosen as the trigger is put on the stack, any player or planeswalker; X is
// twice those counters less 2 (nothing at one counter), and the damage to the
// player and their creatures is one simultaneous instance (Bonfire of the
// Damned's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e1d8c64-4cea-4c2c-8933-be19d13bf07c",
		Name:         "Heart of Bogardan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Heart of Bogardan — cumulative upkeep {2}", "{2}"),
			effects.Targeting(effects.WhenCumulativeUpkeepIsNotPaid("Heart of Bogardan — it deals X damage to target player or planeswalker and each creature that player or that planeswalker's controller controls, where X is twice the number of age counters on it minus 2",
				effects.S82OldB772HeartOfBogardanDamage), effects.TargetPlayerOrPlaneswalker()),
		},
	})
}
