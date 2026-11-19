package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Windcaller Aven — Creature — Bird Wizard {4}{U}{U}, 4/3:
//
//	"Flying
//	 Cycling {U} ({U}, Discard this card: Draw a card.)
//	 When you cycle this card, target creature gains flying until end of
//	 turn."
//
// The cycle trigger fires from the graveyard (CR 702.29c), the shape
// Agonasaur Rex uses. Its target is mandatory: with no creature on the
// battlefield it is removed from the stack, and the card is still drawn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "1da1aa95-b587-4bc3-8bc1-7daa1fc33f00",
		Name:            "Windcaller Aven",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{Cycling("{U}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				InGraveyard(On(game.EventCycle, Self,
					"Windcaller Aven — cycled: target creature gains flying until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							return GrantKeywordUntilEOT{
								Target:   t.ID,
								Keywords: []string{"flying"},
								Label:    "Windcaller Aven — flying until end of turn",
							}.Apply(ctx)
						}
						return nil
					})),
				TargetCreature("target creature"),
			),
		},
	})
}
