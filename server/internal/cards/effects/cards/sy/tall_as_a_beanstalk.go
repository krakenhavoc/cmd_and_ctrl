package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tall as a Beanstalk — Enchantment — Aura {3}{G} (Throne of Eldraine, S82
// old sets):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3, has reach, and is a Giant in addition to
//	 its other types."
//
// Silken Strength's +N/+N and reach, with Puppet Crafting's layer 4 addition:
// Giant is added to the creature types and nothing is removed (CR 205.1b),
// the word read through the Aura's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b638c475-778c-4251-8dac-6a18c3394ae5",
		Name:         "Tall as a Beanstalk",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("reach"),
			{
				Label:     "Attached creature is a Giant in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					giant := source.TextSubtype("Giant")
					if !effects.ZoeticHas(c.Subtypes, giant) {
						c.Subtypes = append(c.Subtypes, giant)
					}
				},
			},
		},
	})
}
