package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Out of Air — Instant {2}{U}{U} (The Lost Caverns of Ixalan, #4470):
//
//	"This spell costs {2} less to cast if it targets a creature spell.
//	 Counter target spell."
//
// Brush Off's shape with a generic reduction: CostsLessIfItTargets reads
// the announced target at CR 601.2f and takes {2} off when it is a
// creature spell, so the discounted price is {U}{U}. The counter is
// Counterspell's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcea912b-58b1-4068-ac4d-1bc89aa7f33c",
		Name:         "Out of Air",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a creature spell.", effects.Creature()),
		},
		OnResolve: effects.CounterTheTargetSpell,
	})
}
