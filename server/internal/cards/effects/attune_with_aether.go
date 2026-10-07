package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "b337c241-f2c4-48e3-a303-371b172fcd18",
		Name:         "Attune with Aether",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1, Energy: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.Controller()
			return SearchLibrary{
				Player:    controller,
				Predicate: IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Attune with Aether — a basic land card",
				Then: func(g *game.Game, _ []uuid.UUID) error {
					return g.AddPlayerCounterByForEffect(controller, controller, game.CounterEnergy, 2)
				},
			}.Apply(ctx)
		},
	})
}
