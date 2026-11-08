package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Welcome to Sweettooth — Enchantment — Saga {1}{G} (Wilds of Eldraine,
// #4471):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Create a 1/1 white Human creature token.
//	 II — Create a Food token.
//	 III — Put X +1/+1 counters on target creature you control, where X is
//	 one plus the number of Foods you control."
//
// The Witch's Vanity's chapter shapes. Chapter III targets, so with no
// creature of yours it is removed rather than put on the stack (CR
// 603.3d). X is counted as the chapter resolves, from the Foods you
// control then (the Saga's own Food included, if it is still there), and
// the counters go on as one placement of X.
//
// No simplification.
func init() {
	const name = "Welcome to Sweettooth"
	Register(Spec{
		OracleID:     "9760c396-fc7d-41bc-8f03-5ee359eaec00",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "create a 1/1 white Human creature token"),
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 white Human"), N: 1}.Apply(NewContext(g, item))
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "create a Food token"),
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
				}),
			ChapterTriggerTargeting(3, SagaChapterLabel(name, 3, "put X +1/+1 counters on target creature you control, where X is one plus the number of Foods you control"),
				TargetCreature("target creature you control", YouControl()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := 1 + len(MatchingBattlefield(ctx, And(YouControl(), Subtype("Food"))))
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
				}),
		},
	})
}
