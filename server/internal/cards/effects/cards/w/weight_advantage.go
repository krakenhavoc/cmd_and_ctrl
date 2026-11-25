package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weight Advantage — Conspiracy (Conspiracy: Take the Crown, #6102):
//
//	"(Start the game with this conspiracy face up in the command zone.)
//	 Each creature you control assigns combat damage equal to its
//	 toughness rather than its power."
//
// A face-up conspiracy (CR 315.5): it starts the game in its owner's
// command zone when it is in their sideboard (ADR 0227), and its static
// works from there. Doran's damage-by-toughness mark, on the creatures its
// controller controls. A creature whose toughness is 0 or less assigns no
// combat damage (CR 510.1a); non-combat damage still uses power.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42bede0a-de6d-4de3-90ea-9d92b0fe8c64",
		Name:         "Weight Advantage",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "Each creature you control assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughness = true
			},
		}},
	})
}
