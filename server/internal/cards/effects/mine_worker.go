package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mine Worker — Artifact Creature — Assembly-Worker {2}, 2/1 (The
// Brothers' War, #5482):
//
//	"{T}: You gain 1 life. If you control creatures named Power Plant
//	 Worker and Tower Worker, you gain 3 life instead."
//
// The ability uses the stack (it is not a mana ability), and the trio is
// counted as it resolves (CR 608.2c): Power Plant Worker and Tower Worker
// are two creatures, one of each name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4de71245-9d9b-44ca-9c13-37dc3dfbbcd5",
		Name:         "Mine Worker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: You gain 1 life. If you control creatures named Power Plant Worker and Tower Worker, you gain 3 life instead.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				n := 1
				if s82BroB11AssemblyTrio(g, item.Controller, "Power Plant Worker", "Tower Worker") {
					n = 3
				}
				return GainLife{Player: item.Controller, Amount: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
