package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forceful Denial — Instant {3}{U}{U} (Commander Legends):
//
//	"Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast it
//	 without paying its mana cost. Put the exiled cards on the bottom in a
//	 random order.)
//	 Counter target spell."
//
// The engine's cascade trigger (Cascade, on the stack above the spell, so it
// resolves first) over Counterspell's body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21d3db6b-58cc-435c-bfd6-8320f6efcbbe",
		Name:         "Forceful Denial",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		Triggered:    []game.TriggeredAbility{effects.Cascade()},
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
