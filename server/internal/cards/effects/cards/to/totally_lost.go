package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Totally Lost — Instant {4}{U} (Ravnica Remastered):
//
//	"Put target nonland permanent on top of its owner's library."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0462a2c-fb88-495f-813b-6476ee3e62bb",
		Name:         "Totally Lost",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve:    effects.GapB79PutTargetOnTopOfItsOwnersLibrary,
	})
}
