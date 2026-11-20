package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cradle to Grave — Instant {1}{B} (Planar Chaos, #5595):
//
//	"Destroy target nonblack creature that entered this turn."
//
// Doom Blade with Gallia's "entered this turn" predicate: the per-object
// entry tally, so a creature that entered, left and came back counts only
// as its new object, and both halves are judged again as it resolves (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a86f9b53-ff44-4b08-a8e9-beef73595de1",
		Name:         "Cradle to Grave",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target nonblack creature that entered this turn", NonBlack(), rfCreatureBEnteredThisTurn()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
