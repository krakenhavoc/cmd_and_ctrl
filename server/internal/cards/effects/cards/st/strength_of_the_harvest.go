package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Strength of the Harvest // Haven of the Harvest — Modal DFC. Front:
// Enchantment — Aura {2}{G/W}: "Enchant creature. Enchanted creature gets
// +1/+1 for each creature and/or enchantment you control." Back: Land
// (mdfc_lands.go).
//
// "You" is the Aura's controller, whoever owns the host. The count is of
// permanents, so a creature that is also an enchantment counts once, and
// the Aura itself is one of the enchantments. Types are the effective ones
// (a permanent turned into a creature counts). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a8c996d-ca93-4c17-ace5-66ecd6b99317",
		Name:         "Strength of the Harvest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, creaturesAndOrEnchantmentsControlledBy),
		},
	})
}

// creaturesAndOrEnchantmentsControlledBy counts "each creature and/or
// enchantment you control", each permanent once.
func creaturesAndOrEnchantmentsControlledBy(g *game.Game, source *game.Card) int {
	return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool {
		for _, t := range c.Effective().Types {
			if t == "Creature" || t == "Enchantment" {
				return true
			}
		}
		return false
	})
}
