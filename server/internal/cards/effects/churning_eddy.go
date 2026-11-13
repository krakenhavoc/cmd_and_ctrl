package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Churning Eddy — Sorcery {3}{U} (Torment, #5612):
//
//	"Return target creature and target land to their owners' hands."
//
// Two clauses, one per printed "target" (#764), each re-checked on its
// own at resolution (CR 608.2b): a target that has left is skipped and
// the other still returns. The survivors go back in one return, Whiplash
// Trap's, so a trigger watching the bounce sees both, and a returned
// commander may stop for its owner's command-zone answer (CR 903.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "774e53b2-5aec-43e1-b4fe-8812b8430401",
		Name:         "Churning Eddy",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			TargetPermanent("target land", Land()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ctx.Game.BounceCardsToHandForEffect(legalTargetCardIDs(ctx))
			return nil
		},
	})
}
