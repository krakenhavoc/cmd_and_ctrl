package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "de8beb92-234b-4684-9185-c7077c8a5133",
		Name:            "Kithkin Spellduster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordPersist},
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}, Sacrifice this creature: Destroy target enchantment.",
			Cost:    Plus(ManaCost("{1}{W}"), SacrificeThis()),
			Targets: TargetPermanent("target enchantment", Enchantment()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return destroyTheTargetPermanent(item, NewContext(g, item))
			},
		}},
	})
}
