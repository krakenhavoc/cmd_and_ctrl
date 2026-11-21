package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hurkyl's Final Meditation — Instant {4}{U}{U}{U} (The Brothers' War
// Commander, #5482):
//
//	"During turns other than yours, this spell costs {3} more to cast.
//	 Return all nonland permanents to their owners' hands. End the turn."
//
// The tax is a self cost increase read as the spell is cast (CR 601.2f):
// the active player at that moment decides it, so it is paid on an
// opponent's turn and not on yours. The two instructions run in printed
// order (ResolveSteps): every nonland permanent, whoever controls it and
// tokens included (a token ceases to exist in hand), returns to its owner's
// hand, then the turn ends as Time Stop's does (CR 724.1): the spell and
// the rest of the stack are exiled (CR 724.1b).
//
// No simplification.
func init() {
	tax := effects.CostsMore(3, "During turns other than yours, this spell costs {3} more to cast.",
		func(q game.CostQuery) bool { return q.Game != nil && !effects.IsYourTurn(q.Game, q.Controller) })
	effects.Register(effects.Spec{
		OracleID:          "a2099033-42d9-47ac-90c7-b41b55b7dadd",
		Name:              "Hurkyl's Final Meditation",
		Completeness:      effects.CompletenessFull,
		Purpose:           game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce}},
		SelfCostModifiers: []game.CostModifier{tax},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BounceAllMatching{Match: effects.Nonland()}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.EndTheTurn{}.Apply(ctx) }),
		),
	})
}
