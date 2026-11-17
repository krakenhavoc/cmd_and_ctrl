package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seed of Hope — Instant {G} (March of the Machine, S82 #5477):
//
//	"Mill two cards. You may put a permanent card from among the milled
//	 cards into your hand. You gain 2 life."
//
// Only cards that actually reached the graveyard count as milled (CR 701.17a,
// 400.7): a card a replacement sent elsewhere is not offered. The life is
// gained after the choice, and also when nothing was milled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "863d6607-912b-4186-9874-87d5473ec481",
		Name:         "Seed of Hope",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return MillToZone{N: 2, ThenTail: ctx.ItemTail(seedOfHopeMilled)}.Apply(ctx)
		},
	})
}
