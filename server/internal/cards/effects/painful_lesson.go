package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Painful Lesson — Sorcery {2}{B}
// (Amonkhet, #5535):
//
//	"Target player draws two cards and loses 2 life."
//
// Sign in Blood's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6be78177-aef0-408c-9769-a1be0775f06d",
		Name:         "Painful Lesson",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2}),
		OnResolve:    signInBloodOnResolve,
	})
}
