package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Terashi's Verdict — Instant — Arcane {1}{W} (Betrayers of Kamigawa, S82
// OLD, #5602):
//
//	"Destroy target attacking creature with power 3 or less."
//
// Kill Shot's attacking-creature target with a power ceiling. Both are
// judged at announce and again on resolution (CR 608.2b): a creature that
// grew past 3 power in response, or stopped attacking, is an illegal
// target and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28ea1f58-3310-4cf1-a834-40f2cc3b6898",
		Name:         "Terashi's Verdict",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking creature with power 3 or less", effects.AttackingCreature(), effects.PowerLE(3)),
		OnResolve:    effects.ResolveSteps(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error { return effects.DestroyTarget{Target: id}.Apply(ctx) })),
	})
}
