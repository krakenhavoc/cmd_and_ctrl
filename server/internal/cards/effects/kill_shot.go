package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kill Shot — Instant {2}{W}:
//
//	"Destroy target attacking creature."
//
// The target is judged against combat at announce and at resolution
// (AttackingCreature, CR 608.2b), as Chastise's is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e1ce32c7-a882-4d3a-9fd6-e5e2943382fb",
		Name:         "Kill Shot",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve:    ResolveSteps(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return DestroyTarget{Target: id}.Apply(ctx) })),
	})
}
