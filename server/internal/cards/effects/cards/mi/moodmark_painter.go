package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moodmark Painter — Creature — Human Shaman {2}{B}{B}, 2/3 (Foundations
// Jumpstart):
//
//	"Undergrowth — When this creature enters, target creature gains
//	 menace and gets +X/+0 until end of turn, where X is the number of
//	 creature cards in your graveyard."
//
// Undergrowth is an ability word: nothing but the count. X is counted as
// the trigger resolves (CR 608.2h) and locked into the pump then (CR
// 611.2c). The target is any creature; a target that left in response
// leaves nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	creatureCards := effects.G2CardsInYourGraveyard(game.Card.IsCreature)
	effects.Register(effects.Spec{
		OracleID:     "f5068b54-b1f8-4751-9a09-fc1b554e8115",
		Name:         "Moodmark Painter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Moodmark Painter — target creature gains menace and gets +X/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"menace"}, Label: "Moodmark Painter — menace"}).Apply(ctx); err != nil {
						return err
					}
					return effects.BoostUntilEOT{Target: t.ID, Power: creatureCards(item, ctx), Label: "Moodmark Painter — +X/+0"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature")),
		},
	})
}
