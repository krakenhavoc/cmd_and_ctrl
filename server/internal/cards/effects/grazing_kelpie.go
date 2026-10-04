package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grazing Kelpie — Creature — Beast {3}{G/U}, 2/3:
//
//	"{G/U}, Sacrifice this creature: Put target card from a graveyard on
//	 the bottom of its owner's library.
//	 Persist"
//
// Persist is PrintedKeywords (#2075).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a2db1035-008e-4de3-b5e5-2a0de83082b2",
		Name:            "Grazing Kelpie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordPersist},
		Activated: []ActivatedAbility{{
			Label:   "{G/U}, Sacrifice this creature: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    Plus(ManaCost("{G/U}"), SacrificeThis()),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					return PutIntoLibrary{Card: t.ID, ToBottom: true}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
