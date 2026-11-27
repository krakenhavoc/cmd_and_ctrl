package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gainsay — Instant {1}{U} (Theros):
//
//	"Counter target blue spell."
//
// Jace's Defeat's blue-spell target clause with Counterspell's counter: the
// target is legal only while the spell is blue (CR 608.2b re-checks it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5115eb6b-efae-4561-912a-f7909cc75775",
		Name:         "Gainsay",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target blue spell", effects.OfColor("U")),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
