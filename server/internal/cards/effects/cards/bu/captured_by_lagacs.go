package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captured by Lagacs — Enchantment — Aura {1}{G}{W} (Modern Horizons 2,
// #5501):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 When this Aura enters, support 2. (Put a +1/+1 counter on each of up
//	 to two target creatures.)"
//
// Cage of Hands' restriction on the host (RestrictAttached) and Together
// Forever's support 2: on a noncreature permanent it is "up to two
// target creatures", with no "other" (CR 701.41a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db1a29f9-1608-4502-a366-89048c2e9bda",
		Name:         "Captured by Lagacs",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersSupport("Captured by Lagacs", 2, false),
		},
	})
}
