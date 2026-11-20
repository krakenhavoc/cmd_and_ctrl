package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skywhaler's Shot — Instant {2}{W} (Kaladesh, #5539):
//
//	"Destroy target creature with power 3 or greater. Scry 1. (Look at
//	 the top card of your library. You may put that card on the
//	 bottom.)"
//
// Big Game Hunter's power bound (judged at announce and as it resolves,
// CR 608.2b), then Voyage's End's scry after the destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3d4e48e-5f98-44a6-ad5c-a1c068aa01eb",
		Name:         "Skywhaler's Shot",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      TargetCreature("target creature with power 3 or greater", PowerGE(3)),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return DestroyTarget{Target: id}.Apply(ctx) }),
			Scry{N: 1},
		),
	})
}
