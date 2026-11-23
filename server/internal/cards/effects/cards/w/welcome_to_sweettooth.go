package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "9760c396-fc7d-41bc-8f03-5ee359eaec00",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "create a 1/1 white Human creature token"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.TokenCard("1/1 white Human"), N: 1}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTrigger(2, effects.SagaChapterLabel(name, 2, "create a Food token"),
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateToken{Controller: item.Controller, Template: effects.FoodToken(), N: 1}.Apply(effects.NewContext(g, item))
				}),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "put X +1/+1 counters on target creature you control, where X is one plus the number of Foods you control"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := 1 + len(effects.MatchingBattlefield(ctx, effects.And(effects.YouControl(), effects.Subtype("Food"))))
					return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: x}.Apply(ctx)
				}),
		},
	})
}
