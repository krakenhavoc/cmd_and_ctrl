package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rescue — Instant {U} (Urza's Destiny, #5621):
//
//	"Return target permanent you control to its owner's hand."
//
// Boomerang narrowed to a permanent you control: it saves your own
// permanent from removal, or resets an enters trigger. A permanent you
// control but don't own goes to its owner's hand. A target gone by
// resolution counters the spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0ae687a-7222-4521-a514-ba10373fed7e",
		Name:         "Rescue",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent you control", YouControl()),
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		OnResolve:    bounceTheTarget,
	})
}
