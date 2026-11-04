package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brush Off — Instant {2}{U}{U}, Secrets of Strixhaven (#4456):
//
//	"This spell costs {1}{U} less to cast if it targets an instant or sorcery
//	 spell.
//	 Counter target spell."
//
// The reduction reads the announced target at CR 601.2f (Price of Fame's
// shape, CostsLessIfItTargets) and takes {1}{U} off the cost by CR 118.7 (a
// blue symbol removes a blue requirement, the generic one a generic), so the
// discounted price is {1}{U}. The counter is Counterspell's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "873694c0-c100-4d24-a74f-1ab26442dd18",
		Name:         "Brush Off",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		SelfCostModifiers: []game.CostModifier{
			socB21CostsLessManaIfItTargets("{1}{U}",
				"This spell costs {1}{U} less to cast if it targets an instant or sorcery spell.",
				Or(Instant(), Sorcery())),
		},
		OnResolve: counterTheTargetSpell,
	})
}
