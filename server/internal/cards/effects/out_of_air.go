package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fcea912b-58b1-4068-ac4d-1bc89aa7f33c",
		Name:         "Out of Air",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a creature spell.", Creature()),
		},
		OnResolve: counterTheTargetSpell,
	})
}
