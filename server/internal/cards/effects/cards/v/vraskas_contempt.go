package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "38ddbe0b-07ef-4cdf-ad6c-cd8b8ba6d206",
		Name:         "Vraska's Contempt",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, LifeGain: 2},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.ExileTarget{Target: id}.Apply(ctx) }),
			effects.GainLife{Amount: 2},
		),
	})
}
