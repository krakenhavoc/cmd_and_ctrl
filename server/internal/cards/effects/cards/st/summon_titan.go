package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Summon: Titan — Enchantment Creature — Saga Giant {3}{G}{G}, 7/7
// (Final Fantasy, #4462):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Mill five cards.
//	 II — Return all land cards from your graveyard to the battlefield
//	 tapped.
//	 III — Until end of turn, another target creature you control gains
//	 trample and gets +X/+X, where X is the number of lands you control.
//	 Reach, trample"
//
// A Saga creature. Chapter II is Splendid Reclamation's body: the lands
// enter together, already tapped. Chapter III is Arahbo's one effect at
// one timestamp, trample and +X/+X, with X read once as it resolves
// (CR 608.2h), so a land that enters later does not grow it.
//
// No simplification.
func init() {
	const name = "Summon: Titan"
	effects.Register(effects.Spec{
		OracleID:        "66ce0c60-f406-4ece-915f-675da4e8b16d",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "mill five cards"), effects.Do(effects.MillCards{N: 5})),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "return all land cards from your graveyard to the battlefield tapped"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.B10ReturnAllLandCardsFromGraveyardTapped(ctx, ctx.Controller())
				}),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "another target creature you control gains trample and gets +X/+X, X the lands you control"),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := effects.CountControlled(g, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
					return effects.UntilEndOfTurn(ctx, id, nil, name+" — trample and +X/+X",
						game.ModifyPTMod(x, x), game.AddKeywordsMod("trample"))
				}),
		},
	})
}
