package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soothing of Sméagol — Instant {1}{U}:
//
//	"Return target nontoken creature to its owner's hand. The Ring
//	 tempts you."
//
// The tempt runs once the creature has reached a hand: a commander may
// stop for its owner's command-zone answer (CR 903.9) and is still on
// the battlefield meanwhile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "35a17b42-9d3f-49ef-8218-2780c5ef5f08",
		Name:         "Soothing of Sméagol",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target nontoken creature", Not(IsTokenPredicate())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			tempt := ringTemptsYouNext(item)
			return ctx.Game.BounceCardsToHandThenForEffect(legalTargetIDs(ctx), func(g *game.Game, _ []uuid.UUID) error {
				return tempt(g)
			})
		},
	})
}
