package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadow of Mortality — Creature — Avatar {13}{B}{B}, 7/7:
//
//	"If your life total is less than your starting life total, this
//	 spell costs {X} less to cast, where X is the difference."
//
// #746: a self cost modifier. The starting life total is the Commander
// format's, game.StartingLife.
func init() {
	Register(Spec{
		OracleID:     "45378874-aa51-4bb6-a161-be336c164778",
		Name:         "Shadow of Mortality",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(LifeBelowStartingTotal(),
				"If your life total is less than your starting life total, this spell costs {X} less to cast, where X is the difference."),
		},
	})
}
