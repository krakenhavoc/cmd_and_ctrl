package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sublime Exhalation — Sorcery {6}{W} (Commander 2016, #5537):
//
//	"Undaunted (This spell costs {1} less to cast for each opponent.)
//	 Destroy all creatures."
//
// Day of Judgment with Curtains' Call's undaunted (CR 702.125a): a self
// cost modifier counting the caster's opponents still in the game. A
// regeneration shield still saves a creature, since the card does not
// say it can't be regenerated.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e0c66e9-7d2b-46bd-890e-886d0c029d1d",
		Name:         "Sublime Exhalation",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(OpponentsOfCaster(), "Undaunted"),
		},
		OnResolve: wrathDestroyAllCreatures,
	})
}
