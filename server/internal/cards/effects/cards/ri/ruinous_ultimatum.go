package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruinous Ultimatum — Sorcery {R}{R}{W}{W}{W}{B}{B} (EDHREC rank 653):
//
//	"Destroy all nonland permanents your opponents control."
//
// The one-sided wrath: seven coloured pips for a board that is
// entirely yours afterwards. Every opponent's nonland permanent —
// creatures, artifacts, enchantments, planeswalkers — leaves as ONE
// event through the S23 DestroyAllMatching primitive, so a Blood
// Artist caught in the sweep sees everything that died beside it
// (CR 700.4). The predicate is the whole card: nonland, and
// controlled by an opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6f38908-aa4f-4f99-a28e-85d11dab52e4",
		Name:         "Ruinous Ultimatum",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, OpponentsOnly: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Nonland(), effects.OpponentControls())}.Apply(ctx)
		},
	})
}
