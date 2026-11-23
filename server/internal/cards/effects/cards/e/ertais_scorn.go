package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ertai's Scorn — Instant {1}{U}{U} (Dominaria United, #5487):
//
//	"This spell costs {U} less to cast if an opponent cast two or more
//	 spells this turn.
//	 Counter target spell."
//
// The reduction removes one blue mana (CR 118.7), or a generic one when the
// cost has no blue left, and is judged when the cost is determined: any one
// opponent of the caster must have cast two or more spells this turn, the
// per-turn tally excluding the spell being priced. The counter is Cancel's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76e6c0e7-593a-4b4a-9eb6-a912f9c3f26f",
		Name:         "Ertai's Scorn",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: a counterspell
		Targets:      effects.TargetSpell("target spell"),
		SelfCostModifiers: []game.CostModifier{
			effects.DmuB02CostsLessMana("{U}", "This spell costs {U} less to cast if an opponent cast two or more spells this turn.",
				effects.DmuB02OpponentCastTwoOrMoreSpells()),
		},
		OnResolve: effects.CounterTheTargetSpell,
	})
}
