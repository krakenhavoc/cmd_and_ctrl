package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hide on the Ceiling — Instant {X}{U}:
//
//	"Exile X target artifacts and/or creatures. Return the exiled cards
//	 to the battlefield under their owners' control at the beginning of
//	 the next end step."
//
// Curse of the Swine's X-counted clause (the number of targets is the
// announced X, CR 601.2c) over Eerie Interlude's delayed return: every
// still-legal target is exiled as one event and the cards that reached
// exile come back together at the next end step, each under its owner's
// control as a new object (CR 400.7). A token that was exiled never
// comes back (CR 111.8). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "376fa374-ead1-4728-ab8b-aca230b72031",
		Name:         "Hide on the Ceiling",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetsCountedByX(effects.TargetPermanent("X target artifacts and/or creatures", effects.Or(effects.Artifact(), effects.Creature()))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTargetsThenScheduleReturn(ctx, "Hide on the Ceiling — return the exiled cards")
		},
	})
}
