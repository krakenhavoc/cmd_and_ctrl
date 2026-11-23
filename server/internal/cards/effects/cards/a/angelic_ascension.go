package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Ascension — Instant {1}{W} (Core Set 2021, #5511):
//
//	"Exile target creature or planeswalker. Its controller creates a 4/4
//	 white Angel creature token with flying."
//
// Resculpt's shape (s82M21B1ExileThenItsControllerCreates): the
// controller is read before the exile moves the card (CR 608.2h), and a
// target that left in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a898412-5b1e-4217-ac78-5b4099ed8e16",
		Name:         "Angelic Ascension",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve:    effects.S82M21B1ExileThenItsControllerCreates("4/4 white Angel with flying"),
	})
}
