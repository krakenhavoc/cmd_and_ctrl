package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Overwhelming Splendor — {6}{W}{W} Enchantment — Aura Curse (Hour of Devastation):
//
//	"Enchant player
//	 Creatures enchanted player controls lose all abilities and have base power
//	 and toughness 1/1.
//	 Enchanted player can't activate abilities that aren't mana abilities or
//	 loyalty abilities."
//
// The two creature statics follow the enchanted player's creatures as they
// come and go (CR 611.3a). The activation ban is a restriction carried by the
// Aura (CR 602.5): it forbids that player's every activation except a mana
// ability (CR 605.1a) or a loyalty ability, from any zone.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1099588c-cc26-4523-8bd2-90e1e1433f53",
		Name:         "Overwhelming Splendor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Static: []game.StaticAbility{
			{
				Label:            "Creatures enchanted player controls lose all abilities.",
				Layer:            game.Layer6Ability,
				RemovesAbilities: true,
				AppliesTo:        effects.OldB239EnchantedPlayersCreatures,
				Apply:            func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
			},
			{
				Label:     "Creatures enchanted player controls have base power and toughness 1/1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: effects.OldB239EnchantedPlayersCreatures,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power = 1
					c.Toughness = 1
				},
			},
		},
		ActivationRestrictions: []game.ActivationRestriction{
			effects.OldB239EnchantedPlayerCantActivateNonMana(
				"Enchanted player can't activate abilities that aren't mana abilities or loyalty abilities."),
		},
	})
}
