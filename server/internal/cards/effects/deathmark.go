package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathmark — Sorcery {B} (Foundations, #4465):
//
//	"Destroy target green or white creature."
//
// Doom Blade with Rending Volley's colour clause: the picker offers
// only a creature that is green or white, and one that has stopped
// being either by resolution is no longer a legal target, so the spell
// does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf09b0af-3cf1-4486-8f65-8cfd2410314a",
		Name:         "Deathmark",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target green or white creature", Or(OfColor("G"), OfColor("W"))),
		OnResolve:    destroyTheTargetPermanent,
	})
}
