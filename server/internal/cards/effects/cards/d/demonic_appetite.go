package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demonic Appetite — Enchantment — Aura {B} (Rise of the Eldrazi, S82 OLD
// batch 484, #5581):
//
//	"Enchant creature you control
//	 Enchanted creature gets +3/+3.
//	 At the beginning of your upkeep, sacrifice a creature."
//
// The sacrifice is the Aura's own trigger and its controller's: they choose
// any creature they control, the enchanted one or not (and none if they
// control none).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52d34007-38c0-4858-8ec8-c622da916277",
		Name:         "Demonic Appetite",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Demonic Appetite — sacrifice a creature",
				func(g *game.Game, item *game.StackItem) error {
					g.PlayerSacrificesForEffect(item.SourceCardID, item.Controller,
						effects.SacrificeSpec("a creature", effects.Creature()), "Demonic Appetite — sacrifice a creature")
					return nil
				}),
		},
	})
}
