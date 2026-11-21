package cardsdi

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dralnu's Crusade — Enchantment {1}{B}{R} (Dominaria Remastered, #5948):
//
//	"All Goblins get +1/+1.
//	 All Goblins are black and are Zombies in addition to their other
//	 creature types."
//
// Three statics on every Goblin on the battlefield, whoever controls it, the
// enchantment's own controller's or not: layer 4 adds the Zombie type
// (CR 613.1d), layer 5 sets the colour to black alone (CR 613.1e, "are
// black" replaces other colours), and layer 7c gives +1/+1.
//
// No simplification.
func init() {
	isGoblin := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.HasSubtype(source.TextSubtype("Goblin"))
	}
	effects.Register(effects.Spec{
		OracleID:     "ff48cf80-4950-4ae4-9f7c-8d826b2f26f7",
		Name:         "Dralnu's Crusade",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "All Goblins are Zombies in addition to their other creature types.",
				Layer:     game.Layer4Type,
				AppliesTo: isGoblin,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					z := source.TextSubtype("Zombie")
					if !slices.Contains(c.Subtypes, z) {
						c.Subtypes = append(c.Subtypes, z)
					}
				},
			},
			{
				Label:     "All Goblins are black.",
				Layer:     game.Layer5Color,
				AppliesTo: isGoblin,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.Colors = []string{source.TextColor("B")}
				},
			},
			{
				Label:     "All Goblins get +1/+1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: isGoblin,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
		},
	})
}
