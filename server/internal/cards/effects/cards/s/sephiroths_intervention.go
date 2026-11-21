package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sephiroth's Intervention — Instant {3}{B} (Final Fantasy, #4462):
//
//	"Destroy target creature. You gain 2 life."
//
// Terminal Criticism's body with an unrestricted target and 2 life. The
// life is gained on resolution whether or not the creature was destroyed
// (an indestructible one stays); a target gone in response fizzles the
// whole spell and nobody gains anything (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94ac2990-58a5-43d4-bdac-ec2f6daf0b51",
		Name:         "Sephiroth's Intervention",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.SocB26DestroyTheTargetAndGainLife(2),
	})
}
