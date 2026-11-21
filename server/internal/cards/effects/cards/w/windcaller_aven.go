package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "1da1aa95-b587-4bc3-8bc1-7daa1fc33f00",
		Name:            "Windcaller Aven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InGraveyard(effects.On(game.EventCycle, effects.Self,
					"Windcaller Aven — cycled: target creature gains flying until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							return effects.GrantKeywordUntilEOT{
								Target:   t.ID,
								Keywords: []string{"flying"},
								Label:    "Windcaller Aven — flying until end of turn",
							}.Apply(ctx)
						}
						return nil
					})),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
