package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Claim the Precious — Sorcery {1}{B}{B}:
//
//	"Destroy target creature. The Ring tempts you."
//
// The tempt runs once the destruction is complete: a commander
// destroyed this way waits for its owner's command-zone answer (CR
// 903.9) on the battlefield, and must not be offered as the
// Ring-bearer meanwhile (ringTemptsYouNext). An indestructible target
// survives and the Ring still tempts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b0a0991-2327-45ea-b47f-c31db367fa1d",
		Name:         "Claim the Precious",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ids := legalTargetIDs(ctx)
			tempt := ringTemptsYouNext(item)
			return ctx.Game.DestroyPermanentsThenForEffect(ids, func(g *game.Game, _ []uuid.UUID) error { return tempt(g) })
		},
	})
}
