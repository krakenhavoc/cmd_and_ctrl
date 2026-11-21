package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kithkin Spellduster — Creature — Kithkin Wizard {4}{W}, 2/3:
//
//	"Flying
//	 {1}{W}, Sacrifice this creature: Destroy target enchantment.
//	 Persist"
//
// Flying and persist are PrintedKeywords (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de8beb92-234b-4684-9185-c7077c8a5133",
		Name:            "Kithkin Spellduster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordPersist},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice this creature: Destroy target enchantment.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
			},
		}},
	})
}
