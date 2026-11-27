package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Opalescence — Enchantment {2}{W}{W} (Urza's Destiny, S82 #5621):
//
//	"Each other non-Aura enchantment is a creature in addition to its other types
//	 and has base power and base toughness each equal to its mana value."
//
// Starfield of Nyx's two layers over the same set, without a controller
// restriction or a count (CR 613.1d, CR 613.4b): the creature type (layer 4) and
// the base size equal to each permanent's mana value (layer 7b), on every other
// non-Aura enchantment, whoever controls it. A token enchantment has mana value
// zero and dies at once (CR 704.5f); a creature that is an Aura (bestow) is spared.
// Two Opalescences each stay an enchantment creature of the other's size, as the
// printed "each other" says.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "59489b46-9d02-4f3c-bcd0-884e7605e9a5",
		Name:         "Opalescence",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Each other non-Aura enchantment is a creature in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: opalescenceApplies,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Types, "Creature") {
						c.Types = append(append([]string(nil), c.Types...), "Creature")
					}
				},
			},
			{
				Label:     "Each such enchantment has base power and base toughness each equal to its mana value.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: opalescenceApplies,
				Apply: func(c *game.Characteristic, target *game.Card, _ *game.Game, _ *game.Card) {
					mv := target.ManaValue()
					c.Power, c.Toughness = mv, mv
				},
			},
		},
	})
}

// opalescenceApplies is the set: every other non-Aura enchantment.
func opalescenceApplies(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.InstanceID != source.InstanceID && target.IsEnchantment() && !target.IsAura()
}
