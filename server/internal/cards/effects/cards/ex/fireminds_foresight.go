package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Firemind's Foresight — Instant {5}{U}{R} (Commander 2015, S82 OLD #5546):
//
//	"Search your library for an instant card with mana value 3, reveal it, and
//	 put it into your hand. Then repeat this process for instant cards with
//	 mana values 2 and 1. Then shuffle."
//
// Three searches, one after another, each continuing from the answer to the
// one before it (a keyed tail), for an instant card of mana value 3, then 2,
// then 1. Each search names a quality, so it may find nothing (CR 701.23b); the
// card is revealed as it goes to hand. The library is shuffled once, after the
// third search, whatever was found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf5765e0-def0-4ddf-9cd7-083feae71660",
		Name:         "Firemind's Foresight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Tutors: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return firemindsForesightSearch(ctx, 3, ctx.Tail(firemindsForesightSecond))
		},
	})
}

// firemindsForesightSearch is one search for an instant card with mana value
// n, revealed into the hand. The last one (n == 1) shuffles; the others hand
// on to `then`.
func firemindsForesightSearch(ctx *effects.Context, n int, then game.KeyedTail) error {
	return effects.SearchLibrary{
		Player: ctx.Controller(),
		Predicate: func(c game.Card) bool {
			return c.IsInstant() && c.ManaValue() == n
		},
		Dest:     game.ZoneHand,
		Limit:    1,
		Reveal:   true,
		Shuffle:  n == 1,
		Reason:   "Firemind's Foresight — search your library for an instant card with mana value " + string(rune('0'+n)),
		ThenTail: then,
	}.Apply(ctx)
}

// The two continuations are registered in init: each refers to the search
// that follows it. On-disk keys: never renamed, never reused.
var firemindsForesightSecond, firemindsForesightThird game.TailRef

func init() {
	firemindsForesightThird = effects.CardTail("then/firemind-s-foresight-third", func(ctx *effects.Context, _ game.TailOutcome) error {
		return firemindsForesightSearch(ctx, 1, game.KeyedTail{})
	})
	firemindsForesightSecond = effects.CardTail("then/firemind-s-foresight-second", func(ctx *effects.Context, _ game.TailOutcome) error {
		return firemindsForesightSearch(ctx, 2, ctx.Tail(firemindsForesightThird))
	})
}
