package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Humble the Brute — Instant {4}{W} (Shadows over Innistrad, #5543):
//
//	"Destroy target creature with power 4 or greater.
//	 Investigate.(Create a Clue token. It's an artifact with "{2},
//	 Sacrifice this token: Draw a card.")"
//
// Stone by Sunlight's "destroy target creature with power 4 or greater",
// then investigate (CR 701.16) in printed order (CR 608.2c), as Auspicious
// Arrival sequences its steps.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bea6606c-ea9a-4199-a8cd-b35d7c47fd7b",
		Name:         "Humble the Brute",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with power 4 or greater", PowerGE(4)),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return DestroyTarget{Target: id}.Apply(ctx) }),
			Investigate{N: 1},
		),
	})
}
