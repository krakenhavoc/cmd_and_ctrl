package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fallow Earth — Sorcery {2}{G} (Mirage, S82 OLD b798, #5633):
//
//	"Put target land on top of its owner's library."
//
// Aethertow's tuck, on a target land: the land goes to the top of its
// OWNER's library, whoever controls it (CR 401.3), through the engine's tuck
// (which sends a commander to the command zone on request).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05a4acfd-3bf3-4ec2-a28a-bc43ba7c6fbf",
		Name:         "Fallow Earth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve:    effects.PutTargetOnTopOfOwnersLibrary,
	})
}
