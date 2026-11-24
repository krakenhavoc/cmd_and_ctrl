package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huatli, the Sun's Heart — Legendary Planeswalker — Huatli {2}{G/W}, loyalty 7:
//
//	"Each creature you control assigns combat damage equal to its toughness
//	 rather than its power.
//	 −3: You gain life equal to the greatest toughness among creatures you
//	 control."
//
// The static is Doran, the Siege Tower's, limited to creatures you control. −3
// reads toughness through the layered view as it resolves; with no creatures
// the greatest toughness is 0 and nothing is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1a13723-a6b1-4160-8eb5-f7a4860898a3",
		Name:            "Huatli, the Sun's Heart",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 7,
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
		Activated: []effects.ActivatedAbility{{
			Label: "−3: You gain life equal to the greatest toughness among creatures you control.",
			Cost:  effects.LoyaltyCost(-3),
			Effect: func(g *game.Game, item *game.StackItem) error {
				greatest := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsCreature() && c.CurrentToughness() > greatest {
						greatest = c.CurrentToughness()
					}
				}
				return effects.GainLife{Amount: greatest}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
