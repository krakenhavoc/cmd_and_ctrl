package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ray of Ruin — Sorcery {4}{B} (The Lost Caverns of Ixalan, #4470):
//
//	"Exile target creature, Vehicle, or nonbasic land. Scry 1."
//
// The target is any permanent that is a creature, has the Vehicle
// subtype, or is a land without the basic supertype; a basic land that
// is not a creature is not a legal target. The spell resolves as two
// steps, the exile and then the scry, and the scry waits for any prompt
// the exile raised (CR 608.2c). A spell whose only target is gone is
// countered whole, scry and all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e36b78ed-4e3e-406d-85b3-fe0deb493418",
		Name:         "Ray of Ruin",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets: effects.TargetPermanent("target creature, Vehicle, or nonbasic land",
			effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"), effects.B10NonbasicLand())),
		OnResolve: effects.ResolveSteps(effects.ExileChosenTarget{}, effects.Scry{N: 1}),
	})
}
