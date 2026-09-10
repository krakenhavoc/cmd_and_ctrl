package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vandalblast — Sorcery {R}:
//
//	"Destroy target artifact you don't control.
//	 Overload {4}{R} (You may cast this spell for its overload cost.
//	 If you do, change 'target' in its text to 'each.')"
//
// S22: overload now works. The alternative-cost machinery carries
// both halves — the {4}{R} price paid instead of the {R}, and the
// deletion of the target clause, which is what turns "target artifact
// you don't control" into "each artifact you don't control".
//
// Note what the deletion buys beyond the sweep: an overloaded
// Vandalblast announces with no targets, so it cannot be fizzled by
// removing "the" artifact in response, and hexproof / shroud on an
// opponent's Darksteel Forge is irrelevant. The predicate survives
// only as the sweep's filter, applied at resolution to whatever is on
// the battlefield then.
//
// The "you don't control" restriction is real and enforced in both
// modes: OpponentControls means your own Sol Ring is never a legal
// target, and never in the overloaded sweep either.
func init() {
	Register(Spec{
		OracleID: "3567c3c8-b3c7-45b7-935b-b1fdbc973720",
		Name:     "Vandalblast",
		Targets:  TargetPermanent("target artifact you don't control", And(Artifact(), OpponentControls())),
		AlternativeCosts: []game.AlternativeCost{
			Overload("{4}{R}"),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.PaidAltCost("overload") {
				// Snapshot before destroying anything: DestroyTarget
				// removes cards from the battlefield slice we would
				// otherwise be ranging over, and a dies-trigger could
				// add one mid-sweep.
				var doomed []uuid.UUID
				for _, c := range ctx.Game.BattlefieldCardsForEffect() {
					if c.IsArtifact() && c.Controller != item.Controller {
						doomed = append(doomed, c.InstanceID)
					}
				}
				for _, id := range doomed {
					if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
