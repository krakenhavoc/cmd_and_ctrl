package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Appetite for the Unnatural — Instant {2}{G} (Kaladesh, #5539):
//
//	"Destroy target artifact or enchantment. You gain 2 life."
//
// One target, then life: a target gone by resolution counters the spell
// whole (CR 608.2b), so no life is gained for nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f041b37e-25af-4aff-b7e3-f07a4d5c6f9a",
		Name:         "Appetite for the Unnatural",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, LifeGain: 2},
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return DestroyTarget{Target: id}.Apply(ctx) }),
			GainLife{Amount: 2},
		),
	})
}
