package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e36b78ed-4e3e-406d-85b3-fe0deb493418",
		Name:         "Ray of Ruin",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Scry: 1},
		Targets: TargetPermanent("target creature, Vehicle, or nonbasic land",
			Or(Creature(), HasSubtype("Vehicle"), b10NonbasicLand())),
		OnResolve: ResolveSteps(ExileChosenTarget{}, Scry{N: 1}),
	})
}
