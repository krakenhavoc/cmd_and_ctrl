package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "e1ce32c7-a882-4d3a-9fd6-e5e2943382fb",
		Name:         "Kill Shot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve:    effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.DestroyTarget{Target: id}.Apply(ctx) })),
	})
}
