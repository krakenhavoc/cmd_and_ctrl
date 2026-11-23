package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Secure the Scene — Sorcery {4}{W} (Core Set 2021, #5511):
//
//	"Exile target nonland permanent. Its controller creates a 1/1 white
//	 Soldier creature token."
//
// Buy Your Silence's shape with a Soldier for the Treasure
// (s82M21B1ExileThenItsControllerCreates): the controller is read before
// the exile (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6a9746c-eba5-4735-aec1-e70edba2c686",
		Name:         "Secure the Scene",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve:    effects.S82M21B1ExileThenItsControllerCreates("1/1 white Soldier"),
	})
}
