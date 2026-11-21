package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galvanic Arc — Enchantment — Aura {2}{R} (Ravnica: City of Guilds, #5600):
//
//	"Enchant creature
//	 When this Aura enters, it deals 3 damage to any target.
//	 Enchanted creature has first strike."
//
// Lashknife's first-strike grant on the host, and Trial of Zeal's enters
// trigger: the Aura itself is the source of the 3 damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72a8620e-1ba2-40a8-9b22-11cc57864adf",
		Name:         "Galvanic Arc",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.GrantToAttached("first strike")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Galvanic Arc — 3 damage to any target", effects.SourceDealsDamageToEachLegalTarget(3)), effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 3))),
		},
	})
}
