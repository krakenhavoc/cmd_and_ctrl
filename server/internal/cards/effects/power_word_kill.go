package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Power Word Kill — Instant {1}{B} (Adventures in the Forgotten Realms):
//
//	"Destroy target non-Angel, non-Demon, non-Devil, non-Dragon creature."
//
// Creature types are read through the layered view, so a creature turned into
// a Dragon is not a legal target and a Dragon turned into a Bear is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea935885-acbc-450d-a3ad-f80c09e12ce7",
		Name:         "Power Word Kill",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: TargetCreature("target non-Angel, non-Demon, non-Devil, non-Dragon creature",
			Not(AnySubtype("Angel", "Demon", "Devil", "Dragon"))),
		OnResolve: destroyTheTargetPermanent,
	})
}
