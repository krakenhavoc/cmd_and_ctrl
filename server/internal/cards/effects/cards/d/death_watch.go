package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Watch — Enchantment — Aura {B} (Visions, S82 older sets batch 790,
// #5631):
//
//	"Enchant creature
//	 When enchanted creature dies, its controller loses life equal to its power
//	 and you gain life equal to its toughness."
//
// Banewasp Affliction's dies trigger. Power and toughness are the creature's
// last-known values as it died (CR 608.2h); the life is lost by the player who
// controlled it as it died, and gained by the Aura's controller. A creature with
// no positive power or toughness moves no life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f5b6cba-3b57-4e6a-afa7-0cb62e5cc7d4",
		Name:         "Death Watch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Death Watch — its controller loses life equal to its power and you gain life equal to its toughness",
				effects.Old790DeathWatch),
		},
	})
}
