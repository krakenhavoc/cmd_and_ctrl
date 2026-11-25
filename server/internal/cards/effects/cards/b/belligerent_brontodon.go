package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Belligerent Brontodon — Creature — Dinosaur {5}{G}{W}, 4/6 (Ixalan,
// #5532):
//
//	"Each creature you control assigns combat damage equal to its
//	 toughness rather than its power."
//
// Gauntlets of Light's layer 6 flag (CR 510.1c), on every creature its
// controller controls, the Brontodon included, whether or not toughness
// is the greater. Only combat damage assignment changes; the creatures'
// power is still their power.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa2d8440-a44a-44fe-8624-0e8b4be53f44",
		Name:         "Belligerent Brontodon",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Each creature you control assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughnessAlways = true
			},
		}},
	})
}
