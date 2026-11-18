package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// True Polymorph — Instant {4}{U}{U} (Adventures in the Forgotten Realms):
//
//	"Target artifact or creature becomes a copy of another target artifact or
//	 creature."
//
// Two clauses in printed order: the permanent that changes, then a DIFFERENT
// artifact or creature to copy ("another", so the second clause is Distinct).
// The effect states no duration, so it lasts indefinitely (CR 611.2a): a
// BecomeCopy with CopyIndefinite (Unstable Shapeshifter's shape). The copiable
// values are read as the spell resolves (CR 707.2). If either target is illegal
// on resolution the spell does nothing: there is nothing to become, or nothing
// to copy (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d24e181-74c1-4913-a18a-59378ec7d9e7",
		Name:         "True Polymorph",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
			Distinct(TargetPermanent("another target artifact or creature", Or(Artifact(), Creature()))),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			becomes, ok := ctx.ClauseTarget(0)
			if !ok || !ctx.IsTargetLegal(becomes) {
				return nil
			}
			model, ok := ctx.ClauseTarget(1)
			if !ok || !ctx.IsTargetLegal(model) {
				return nil
			}
			return BecomeCopy{
				Targets:  []uuid.UUID{becomes.ID},
				Of:       model.ID,
				Duration: CopyIndefinite,
				Label:    "True Polymorph — becomes a copy",
			}.Apply(ctx)
		},
	})
}
