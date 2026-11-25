package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deep Freeze — Enchantment — Aura {2}{U} (Dominaria, S82 OLD, #5530):
//
//	"Enchant creature
//	 Enchanted creature has base power and toughness 0/4, has defender, loses
//	 all other abilities, and is a blue Wall in addition to its other colors
//	 and types."
//
// One continuous effect in four layers (Blade of the Oni's "in addition to"
// shape with Spider-Man No More's ability loss): the Wall subtype in layer 4
// and blue in layer 5, both added (CR 205.1b), layer 6 removing every ability
// but the defender it grants, and base 0/4 in layer 7b so counters and anthems
// still apply on top. Each part continues if the Aura loses its abilities
// part-way through the pass (CR 613.6). "Wall" and "blue" are words of the
// Aura's text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "648aead4-dc0b-47bc-911c-8dcd2cfd32d3",
		Name:         "Deep Freeze",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			{
				Label:                 "Enchanted creature is a Wall in addition to its other types.",
				Layer:                 game.Layer4Type,
				ContinuesAfterRemoval: true,
				AppliesTo:             effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					if wall := source.TextSubtype("Wall"); !effects.HasFold(c.Subtypes, wall) {
						c.Subtypes = append(c.Subtypes, wall)
					}
				},
			},
			{
				Label:                 "Enchanted creature is blue in addition to its other colors.",
				Layer:                 game.Layer5Color,
				ContinuesAfterRemoval: true,
				AppliesTo:             effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					if blue := source.TextColor("U"); !effects.HasFold(c.Colors, blue) {
						c.Colors = append(c.Colors, blue)
					}
				},
			},
			effects.LoseAllAbilities("defender"),
			effects.SetAttachedBasePT(0, 4),
		},
	})
}
