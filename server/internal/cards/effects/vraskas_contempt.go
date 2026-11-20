package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vraska's Contempt — Instant {2}{B}{B} (Ixalan, #5532):
//
//	"Exile target creature or planeswalker. You gain 2 life."
//
// Extended Absence's exile, then the life gain in printed order
// (CR 608.2c). A target that has become illegal counters the spell whole
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "38ddbe0b-07ef-4cdf-ad6c-cd8b8ba6d206",
		Name:         "Vraska's Contempt",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, LifeGain: 2},
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error { return ExileTarget{Target: id}.Apply(ctx) }),
			GainLife{Amount: 2},
		),
	})
}
