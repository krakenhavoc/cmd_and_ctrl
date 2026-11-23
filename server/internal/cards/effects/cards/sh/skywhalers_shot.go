package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "f3d4e48e-5f98-44a6-ad5c-a1c068aa01eb",
		Name:         "Skywhaler's Shot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets:      effects.TargetCreature("target creature with power 3 or greater", effects.PowerGE(3)),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.DestroyTarget{Target: id}.Apply(ctx) }),
			effects.Scry{N: 1},
		),
	})
}
