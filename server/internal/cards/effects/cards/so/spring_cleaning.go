package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spring Cleaning — Instant {1}{G} (Lorwyn, #5593):
//
//	"Destroy target enchantment. Clash with an opponent. If you win,
//	 destroy all enchantments your opponents control. (Each clashing
//	 player reveals the top card of their library, then puts that card on
//	 their choice of the top or bottom. A player wins if their card had a
//	 greater mana value.)"
//
// Destroy the target, then clash. On a win every enchantment an opponent
// controls is destroyed as one sweep (mass destroy, CR 701.8), the target
// included if it is theirs and still there. A target that became illegal
// does nothing, clash included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b3b81cf-f020-4e1e-95c7-db44ca153dc1",
		Name:         "Spring Cleaning",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepEnchantments, How: game.SweepDestroy, OpponentsOnly: true, Partial: true}},
		Targets:      effects.TargetPermanent("target enchantment", effects.Enchantment()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if _, ok := effects.B16FirstLegalTargetCard(ctx); !ok {
				return nil
			}
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.ClashWithAnOpponent{Then: ctx.ItemTail(springCleaningClashed)}.Apply(ctx)
		},
	})
}

// springCleaningClashed is "If you win, destroy all enchantments your
// opponents control". An on-disk key.
var springCleaningClashed = effects.CardTail("then/spring-cleaning-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) {
		return nil
	}
	return effects.DestroyAllMatching{Match: effects.And(effects.Enchantment(), effects.OpponentControls())}.Apply(ctx)
})
