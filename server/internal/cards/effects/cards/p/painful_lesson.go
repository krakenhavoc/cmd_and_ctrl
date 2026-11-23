package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Painful Lesson — Sorcery {2}{B}
// (Amonkhet, #5535):
//
//	"Target player draws two cards and loses 2 life."
//
// Sign in Blood's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6be78177-aef0-408c-9769-a1be0775f06d",
		Name:         "Painful Lesson",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, LifeLoss: 2}),
		OnResolve:    effects.SignInBloodOnResolve,
	})
}
