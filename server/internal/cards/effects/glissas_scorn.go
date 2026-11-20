package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glissa's Scorn — Instant {1}{G} (New Phyrexia, #5575):
//
//	"Destroy target artifact. Its controller loses 1 life."
//
// The artifact's controller is read before the destroy, so the life loss
// still lands when it is indestructible or regenerated
// (nphThenItsControllerLosesOne). A target gone by resolution does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f30f17f0-bd2d-4f10-9739-111917acc100",
		Name:         "Glissa's Scorn",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve:    nphThenItsControllerLosesOne(func(ctx *Context, id uuid.UUID) error { return DestroyTarget{Target: id}.Apply(ctx) }),
	})
}
