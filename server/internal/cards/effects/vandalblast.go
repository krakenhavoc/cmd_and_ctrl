package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vandalblast — Sorcery {R}:
//
//	"Destroy target artifact you don't control.
//	 Overload {4}{R} (You may cast this spell for its overload cost.
//	 If you do, change "target" in its text to "each.")"
//
// S22 landed overload. The alternative-cost machinery carries both
// halves — the {4}{R} price paid instead of the {R}, and the deletion
// of the target clause, which is what turns "target artifact you
// don't control" into "each artifact you don't control".
//
// Note what the deletion buys beyond the sweep: an overloaded
// Vandalblast announces with no targets, so it cannot be fizzled by
// removing "the" artifact in response, and hexproof / shroud on an
// opponent's Darksteel Forge is irrelevant. The predicate survives
// only as the sweep's filter, applied at resolution to whatever is on
// the battlefield then.
//
// S23: that filter is now literally the same predicate the target
// clause is built from — And(Artifact(), OpponentControls()) — rather
// than a hand-inlined copy of it, which is the whole point of
// building the mass primitives on the S20 predicate library. "You
// don't control" is real and enforced in both modes: your own Sol
// Ring is never a legal target, and never in the sweep either.
func init() {
	artifactYouDontControl := And(Artifact(), OpponentControls())
	Register(Spec{
		OracleID: "3567c3c8-b3c7-45b7-935b-b1fdbc973720",
		Name:     "Vandalblast",
		Targets:  TargetPermanent("target artifact you don't control", artifactYouDontControl),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{4}{R}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				return DestroyAllMatching{Match: artifactYouDontControl}.Apply(ctx)
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
