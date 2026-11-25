package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thorn Elemental — Creature — Elemental {5}{G}{G}, 7/7 (Dominaria, S82 OLD,
// #5530):
//
//	"You may have this creature assign its combat damage as though it weren't
//	 blocked."
//
// Pride of Lions' static (CR 510.1c, Spinebiter's file): when blocked, its
// controller may assign all its combat damage to the player or planeswalker it
// attacks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9cdf840d-314e-46ef-9c04-74f145d9e65b",
		Name:         "Thorn Elemental",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "You may have this creature assign its combat damage as though it weren't blocked.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.AssignsAsThoughUnblocked = true
			},
		}},
	})
}
