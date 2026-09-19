package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rite of Replication — "Kicker {5}. Create a token that's a copy of
// target creature. If this spell was kicked, create five of those
// tokens instead."
//
// The kicker whose payoff is a NUMBER rather than a branch, and the
// card on #664's waiting list from batch issue #300. Same read as
// Burst Lightning — `ctx.WasKicked()` off the paid record — feeding
// CreateTokenCopy's N.
//
// "Five of THOSE tokens": one template, taken once from the target
// (CR 707.2 copiable values), then five tokens from it. Building the
// template per token would read the target five times and could
// diverge if something changed it mid-resolution; CreateTokenCopy
// already takes N for exactly this reason.
//
// SANDBOX SIMPLIFICATION, and it is the printed card's most-asked
// question: the kicked tokens enter under THIS spell's controller,
// which is right, but the engine has no "gain control" clause here to
// get wrong — Rite of Replication's tokens are always its caster's.
// Nothing is deferred.
func init() {
	Register(Spec{
		OracleID:     "fb60739e-1dc3-481d-a056-ad72e665c680",
		Name:         "Rite of Replication",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{
			Kicker("{5}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 1
			if ctx.WasKicked() {
				n = 5
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return CreateTokenCopy{
					Controller: item.Controller,
					Copy:       t.ID,
					N:          n,
				}.Apply(ctx)
			}
			return nil
		},
	})
}
