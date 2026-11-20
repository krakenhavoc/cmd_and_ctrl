package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Victorious Destruction — Sorcery {4}{R} (New Phyrexia, #5575):
//
//	"Destroy target artifact or land. Its controller loses 1 life."
//
// Glissa's Scorn's body over an artifact or a land: the controller is read
// before the destroy (nphThenItsControllerLosesOne), and a target gone by
// resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa5d6c02-53b2-4144-87cb-e223b529ba7d",
		Name:         "Victorious Destruction",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or land", Or(Artifact(), Land())),
		OnResolve:    nphThenItsControllerLosesOne(func(ctx *Context, id uuid.UUID) error { return DestroyTarget{Target: id}.Apply(ctx) }),
	})
}
