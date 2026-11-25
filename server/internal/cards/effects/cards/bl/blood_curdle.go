package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Curdle — Instant {3}{B} (Ikoria: Lair of Behemoths, #5512):
//
//	"Destroy target creature. Put a menace counter on a creature you control."
//
// The counter's creature is chosen on resolution, not targeted, from the
// creatures you control then (the destroyed one is gone), so a creature the
// destroy just removed is not a choice and the counter is put on whichever
// one the controller picks; with exactly one it goes there without a
// question. With none it does nothing. The destroy is skipped alone if its
// target is illegal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8b50de69-121a-48c4-bd7c-c8858f35f6ed",
		Name:         "Blood Curdle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.DestroyTarget{Target: id}.Apply(ctx) }),
			effects.S82OldB113CounterOnACreatureYouControl("Blood Curdle", game.CounterMenace),
		),
	})
}
