package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vapor Snag — Instant {U} (New Phyrexia, #5575):
//
//	"Return target creature to its owner's hand. Its controller loses 1 life."
//
// The creature's controller is read before it leaves, so the life loss
// lands on the player who controlled it (nphThenItsControllerLosesOne).
// A target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1408bae-ed5f-4837-a1a7-438dc30a2786",
		Name:         "Vapor Snag",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.NphThenItsControllerLosesOne(func(ctx *effects.Context, id uuid.UUID) error { return effects.BounceToHand{Target: id}.Apply(ctx) }),
	})
}
