package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Waking Nightmare — Sorcery — Arcane {2}{B} (Champions of Kamigawa,
// #5603):
//
//	"Target player discards two cards."
//
// Mind Rot's text and body: the target player chooses the two cards (CR
// 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3f34ef2-1255-4da2-9f29-8c79fee7a2c5",
		Name:         "Waking Nightmare",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.MindRotOnResolve,
	})
}
