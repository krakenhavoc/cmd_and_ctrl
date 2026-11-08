package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "66ce0c60-f406-4ece-915f-675da4e8b16d",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "mill five cards"), Do(MillCards{N: 5})),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "return all land cards from your graveyard to the battlefield tapped"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return b10ReturnAllLandCardsFromGraveyardTapped(ctx, ctx.Controller())
				}),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "another target creature you control gains trample and gets +X/+X, X the lands you control"),
				Another(TargetCreature("another target creature you control", YouControl())),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := countControlled(g, ctx.Controller(), func(c game.Card) bool { return c.IsLand() })
					return untilEndOfTurn(ctx, id, nil, name+" — trample and +X/+X",
						game.ModifyPTMod(x, x), game.AddKeywordsMod("trample"))
				}),
		},
	})
}
