package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Power Word Kill — Instant {1}{B} (Adventures in the Forgotten Realms):
//
//	"Destroy target non-Angel, non-Demon, non-Devil, non-Dragon creature."
//
// Creature types are read through the layered view, so a creature turned into
// a Dragon is not a legal target and a Dragon turned into a Bear is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea935885-acbc-450d-a3ad-f80c09e12ce7",
		Name:         "Power Word Kill",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetCreature("target non-Angel, non-Demon, non-Devil, non-Dragon creature",
			effects.Not(effects.AnySubtype("Angel", "Demon", "Devil", "Dragon"))),
		OnResolve: effects.DestroyTheTargetPermanent,
	})
}
