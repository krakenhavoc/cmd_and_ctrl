package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Solid Footing — Enchantment — Aura {W} (Ikoria: Lair of Behemoths, S82 OLD
// batch 120, #5512):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+1.
//	 As long as enchanted creature has vigilance, it assigns combat damage
//	 equal to its toughness rather than its power."
//
// The pump is PumpAttached. The damage rule is Felothar's
// DamageByToughnessAlways flag (CR 510.1c, toughness whether or not it is
// the greater), set by a layer 6 static on the enchanted creature only while
// it has vigilance, read through the layers so a vigilance granted by
// another effect counts and one lost does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "edc91e1a-a305-4c14-9d85-077a410528fa",
		Name:            "Solid Footing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			{
				Layer: game.Layer6Ability,
				Label: "As long as enchanted creature has vigilance, it assigns combat damage equal to its toughness rather than its power.",
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && game.HasKeyword(target, "vigilance")
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.DamageByToughnessAlways = true
				},
			},
		},
	})
}
