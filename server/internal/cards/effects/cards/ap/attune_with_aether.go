package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Attune with Aether — Sorcery {G}:
//
//	"Search your library for a basic land card, reveal it, put it into
//	 your hand, then shuffle. You get {E}{E} (two energy counters)."
//
// ADR 0129 PR 1. The energy comes after the search, whether or not a
// land was found.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b337c241-f2c4-48e3-a303-371b172fcd18",
		Name:         "Attune with Aether",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Energy: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			controller := ctx.Controller()
			return effects.SearchLibrary{
				Player:    controller,
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Attune with Aether — a basic land card",
				ThenTail:  ctx.TailN(effects.ThenGetEnergy, 2),
			}.Apply(ctx)
		},
	})
}
