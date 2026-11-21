package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harmless Assault — Instant {2}{W}{W}:
//
//	"Prevent all combat damage that would be dealt this turn by
//	 attacking creatures."
//
// #2026's combat status, read as each creature would deal combat
// damage (CR 609.7b): an attacking creature's combat damage is
// prevented, and a blocker's is not (the ruling).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bfcc3a16-1ca8-4112-a10e-d4e52d3daa8d",
		Name:         "Harmless Assault",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.SourceShieldSpell(effects.CombatShieldAgainstCreatures(game.DamageSourceFilter{Combat: game.SourceCombatAttacking})),
	})
}
