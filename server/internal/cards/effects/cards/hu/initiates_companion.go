package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Initiate's Companion — Creature — Cat {1}{G}, 3/1 (Jumpstart, #5510):
//
//	"Whenever this creature deals combat damage to a player, untap target
//	 creature or land."
//
// Civic Gardener's untap over creatures and lands, on a combat-damage
// trigger. The target is chosen as the trigger goes on the stack (CR
// 603.3d) and may be the Companion itself; one that is gone or no longer a
// creature or land as it resolves is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3264224-43a4-4c6b-b363-2d147687dfc1",
		Name:         "Initiate's Companion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Initiate's Companion — untap target creature or land", effects.UntapTheTarget),
				effects.TargetPermanent("target creature or land", effects.Or(effects.Creature(), effects.Land()))),
		},
	})
}
