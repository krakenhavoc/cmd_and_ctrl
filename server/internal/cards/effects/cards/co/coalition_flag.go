package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coalition Flag — Enchantment — Aura {W} (Apocalypse, S82 OLD, #5614):
//
//	"Enchant creature you control
//	 Enchanted creature is a Flagbearer.
//	 While an opponent is choosing targets as part of casting a spell they
//	 control or activating an ability they control, that player must choose at
//	 least one Flagbearer on the battlefield if able."
//
// The Flagbearer seam (#6236). "Is a Flagbearer" sets the creature's types to
// Flagbearer (Eye of Nidhogg's layer 4 shape, CR 205.1a: no "in addition"
// keeps none of the old types), and the targeting rule rides the enchanted
// creature as a layer 6 mark the engine binds at the cast and activation gates
// (FlagbearerTargetRule's mark on the permanent that the Aura makes a
// Flagbearer).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b45a0a2f-39fd-47b9-bc9d-3d4321b377d3",
		Name:         "Coalition Flag",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Static: []game.StaticAbility{
			{
				Label:                 "Enchanted creature is a Flagbearer.",
				Layer:                 game.Layer4Type,
				ContinuesAfterRemoval: true,
				AppliesTo:             effects.AttachedToSource,
				// "Flagbearer" is a word of the Aura's text, read through it
				// (CR 612.2).
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
					c.SetSubtypes(source.TextSubtypes("Flagbearer"))
				},
			},
			{
				Label:     "While an opponent is choosing targets as part of casting a spell they control or activating an ability they control, that player must choose at least one Flagbearer on the battlefield if able.",
				Layer:     game.Layer6Ability,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.FlagbearerTargetRule = true
				},
			},
		},
	})
}
