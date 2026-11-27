package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashling, the Extinguisher — Legendary Creature — Elemental Shaman {2}{B}{B},
// 4/4 (Eventide, S82 OLD batch 520, #5590):
//
//	"Whenever Ashling deals combat damage to a player, choose target creature that
//	 player controls. The player sacrifices that creature."
//
// Blind Zealot's trigger: the target is chosen as the trigger goes on the stack,
// among the creatures the damaged player controls (ControlledByThatPlayer, read off
// the damage event), and as it resolves that player sacrifices it if it is still a
// legal target (CR 608.2b), the sacrifice being theirs and not a destruction, so
// regeneration and indestructible do not stop it. The trigger exists only when a
// creature can be chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e89e102f-e3bc-4c16-bb27-09db83cec5ec",
		Name:         "Ashling, the Extinguisher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisDealsCombatDamageToAPlayer("Ashling, the Extinguisher — choose target creature that player controls; the player sacrifices that creature",
				func(g *game.Game, item *game.StackItem) error {
					id, ok := effects.B16FirstLegalTargetCard(effects.NewContext(g, item))
					if !ok {
						return nil
					}
					return g.SacrificePermanentForEffect(id)
				}),
				effects.ControlledByThatPlayer(effects.TargetCreature("target creature that player controls"))),
		},
	})
}
