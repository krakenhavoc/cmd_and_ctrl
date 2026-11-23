package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "774e53b2-5aec-43e1-b4fe-8812b8430401",
		Name:         "Churning Eddy",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPermanent("target land", effects.Land()),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ctx.Game.BounceCardsToHandForEffect(effects.LegalTargetCardIDs(ctx))
			return nil
		},
	})
}
