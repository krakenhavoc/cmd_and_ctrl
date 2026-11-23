package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "cf09b0af-3cf1-4486-8f65-8cfd2410314a",
		Name:         "Deathmark",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target green or white creature", effects.Or(effects.OfColor("G"), effects.OfColor("W"))),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
