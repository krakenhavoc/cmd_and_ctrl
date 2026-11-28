package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seizures — Enchantment — Aura {1}{B} (Ice Age, S82 OLD #5636):
//
//	"Enchant creature
//	 Whenever enchanted creature becomes tapped, this Aura deals 3 damage to that
//	 creature's controller unless that player pays {3}."
//
// Corrupted Roots' tap watcher, with a pay-unless in place of the life loss.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "550921f5-ec65-460d-9ada-f2c5a4c4b2c4",
		Name:         "Seizures",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
			AppliesTo: effects.EnchantedPermanentBecameTapped,
			Key:       "Seizures — 3 damage to its controller unless they pay {3}",
			Effect:    effects.DamageEnchantedControllerUnlessPayB829("Seizures", 3, "{3}"),
		}},
	})
}
