package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doran, the Siege Tower — Legendary Creature — Treefolk Shaman {W}{B}{G},
// 0/5 (Lorwyn, #5593):
//
//	"Each creature assigns combat damage equal to its toughness rather
//	 than its power."
//
// Ancient Lumberknot's damage-by-toughness mark, on every creature on the
// battlefield, anyone's, with no condition: the rule CR 510.1a reads power
// and Doran makes the whole table read toughness. A creature whose
// toughness is 0 or less assigns no combat damage (CR 510.1a). Non-combat
// damage still uses power.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a46d96d9-8e59-4777-8208-7730b9e33240",
		Name:         "Doran, the Siege Tower",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Each creature assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughness = true
			},
		}},
	})
}
