package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Broken Concentration — Instant {1}{U}{U} (Shadows over Innistrad, #5543):
//
//	"Counter target spell.
//	 Madness {3}{U} (If you discard this card, discard it into exile.
//	 When you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Cancel's counter (counterTheTargetSpell). Madness is grown by the engine
// from the one string (CR 702.35a, #657).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b0fa80e-6b16-4a4a-9776-6a91f145fc63",
		Name:         "Broken Concentration",
		Completeness: effects.CompletenessFull,
		Madness:      "{3}{U}",
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
