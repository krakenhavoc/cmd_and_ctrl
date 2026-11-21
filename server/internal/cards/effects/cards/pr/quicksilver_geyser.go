package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quicksilver Geyser — Instant {4}{U} (Mirrodin Besieged,
// #5576):
//
//	"Return up to two target nonland permanents to their owners' hands."
//
// This Town Ain't Big Enough's "up to two target nonland permanents" bounce
// without its discount (b35BounceChosenPermanents): every chosen permanent
// still legal as it resolves returns at once (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ee7b00b-b875-48e2-92f2-889c154b0f82",
		Name:         "Quicksilver Geyser",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("up to two target nonland permanents", effects.Nonland()).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B35BounceChosenPermanents(ctx)
		},
	})
}
