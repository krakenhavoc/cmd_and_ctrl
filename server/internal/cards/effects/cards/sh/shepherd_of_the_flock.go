package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shepherd of the Flock // Usher to Safety — Creature — Human Peasant {1}{W}, 3/1 // Instant — Adventure {W} (Throne of Eldraine, #5516):
//
//	Shepherd of the Flock — no rules text.
//	Usher to Safety — "Return target permanent you control to its owner's hand."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// Usher to Safety is Rescue's text.
//
// No simplification.
const shepherdOfTheFlockOracleID = "6d611800-5b0e-40ca-82a2-5b294a1ebf6d"

func init() {
	effects.Register(effects.Spec{
		OracleID:     shepherdOfTheFlockOracleID,
		Name:         "Shepherd of the Flock",
		Completeness: effects.CompletenessFull,
	})
	effects.Register(effects.Spec{
		OracleID:     shepherdOfTheFlockOracleID + "#1",
		Name:         "Usher to Safety",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		Targets:      effects.TargetPermanent("target permanent you control", effects.YouControl()),
		OnResolve:    effects.BounceTheTarget,
	})
}
