package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hindervines — Instant {2}{G}:
//
//	"Prevent all combat damage that would be dealt this turn by
//	 creatures with no +1/+1 counters on them."
//
// #2026's counter test, read as each creature would deal combat damage
// (CR 609.7b; the ruling: "Hindervines checks whether a creature has a
// +1/+1 counter on it at the moment it deals damage").
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d13fabc-1e0f-41f2-8873-9f44b68f7e43",
		Name:         "Hindervines",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.SourceShieldSpell(effects.CombatShieldAgainstCreatures(game.DamageSourceFilter{NoCounters: []string{game.CounterPlusOne}})),
	})
}
