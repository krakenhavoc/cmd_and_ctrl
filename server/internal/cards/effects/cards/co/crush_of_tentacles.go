package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crush of Tentacles — Sorcery {4}{U}{U} (Oath of the Gatewatch, #5545):
//
//	"Surge {3}{U}{U} (You may cast this spell for its surge cost if you or a
//	 teammate has cast another spell this turn.)
//	 Return all nonland permanents to their owners' hands. If this spell's
//	 surge cost was paid, create an 8/8 blue Octopus creature token."
//
// The bounce is symmetric and takes tokens with it (they cease to exist); the
// Octopus is made after, so it stays. "If this spell's surge cost was paid"
// reads the cost announced with the cast (CR 601.2b, CR 702.117a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "4ede153c-89cb-4121-bf23-a1d213d6f238",
		Name:             "Crush of Tentacles",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Surge("{3}{U}{U}")},
		// ADR 0126 §6: every nonland permanent, ours too.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BounceAllMatching{Match: effects.Nonland()}).Apply(ctx); err != nil {
				return err
			}
			if !effects.SurgeCostWasPaid(ctx) {
				return nil
			}
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("8/8 blue Octopus"), N: 1}.Apply(ctx)
		},
	})
}
