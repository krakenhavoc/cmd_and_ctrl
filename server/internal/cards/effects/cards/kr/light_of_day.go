package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Light of Day — Enchantment {3}{W} (Tempest, S82 OLD batch 765, #5628):
//
//	"Black creatures can't attack or block."
//
// A layer 6 restriction over every black creature, whoever controls it. The
// colour is a word of this card's text (CR 612.2), read after layer 5, so a
// creature made black attacks no more and one made another colour may.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cd854e00-c9da-4593-8402-4bae637f5185",
		Name:         "Light of Day",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Black creatures can't attack or block.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.HasColor(source.TextColor("B"))
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
	})
}
