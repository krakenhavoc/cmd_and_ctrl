package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wirecat — Artifact Creature — Cat {4}, 4/3 (Urza's Saga, S82 OLD batch 732,
// #5624):
//
//	"This creature can't attack or block if an enchantment is on the
//	 battlefield."
//
// A layer 6 restriction on itself (CR 613.1f) that applies while any
// enchantment, whoever controls it, is on the battlefield, read through the
// layers each time attackers or blockers are declared. The restriction goes
// the moment the last enchantment does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57828973-df3d-4288-988d-147a3016120e",
		Name:         "Wirecat",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "This creature can't attack or block if an enchantment is on the battlefield.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				if !effects.SelfOnly(target, g, source) {
					return false
				}
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsEnchantment() {
						return true
					}
				}
				return false
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
	})
}
