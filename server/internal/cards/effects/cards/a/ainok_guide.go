package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ainok Guide — Creature — Dog Scout {1}{G}, 1/1 (Foundations Jumpstart):
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature.
//	 • Search your library for a basic land card, reveal it, then shuffle and
//	   put that card on top."
//
// A modal enters trigger (CR 603.3c), Adaptive Sporesinger's shape. The
// search is Campus Guide's, not optional here, though a library without a
// basic land finds nothing.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Ainok Guide — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on this creature.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		effects.ModeDoing("Search your library for a basic land card, reveal it, then shuffle and put that card on top.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneLibrary,
					ToTop:     true,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Ainok Guide — a basic land card",
				}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "e5f96482-4bae-4d1b-97b2-dcc09676eb6e",
		Name:         "Ainok Guide",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered:    []game.TriggeredAbility{enters},
	})
}
