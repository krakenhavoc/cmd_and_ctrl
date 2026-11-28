package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rushing River — Instant {2}{U} (Planeshift, S82 OLD b641, #5615):
//
//	"Kicker—Sacrifice a land. (You may sacrifice a land in addition to any
//	 other costs as you cast this spell.)
//	 Return target nonland permanent to its owner's hand. If this spell was
//	 kicked, return another target nonland permanent to its owner's hand."
//
// The kicker is a sacrifice paid as the spell is cast (CR 702.33a, 601.2h). The
// second target exists only when it was paid (CR 702.33d), and "another" makes it
// a different object from the first. Each clause is judged for itself as the
// spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	first := effects.TargetPermanent("target nonland permanent", effects.Nonland())
	kicker := effects.KickerSacrifice("a land", effects.Land())
	kicker.Targets = effects.Clauses(first, effects.Distinct(effects.TargetPermanent("another target nonland permanent", effects.Nonland())))
	effects.Register(effects.Spec{
		OracleID:      "6e2a695d-305a-4bbe-b302-11fdc5c2a451",
		Name:          "Rushing River",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OptionalCosts: []game.AdditionalCost{kicker},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.OldB641BounceClause(ctx, 0) }),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.OldB641BounceClause(ctx, 1)
			}),
		),
	})
}
