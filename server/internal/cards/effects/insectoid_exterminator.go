package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insectoid Exterminator — {2}{B} Creature — Insect Mutant 2/2 (Teenage
// Mutant Ninja Turtles, #4457):
//
//	"Flying
//	 Disappear — At the beginning of your end step, if a permanent left
//	 the battlefield under your control this turn, scry 1."
//
// Disappear is an ability word (CR 207.2c) over revolt's condition: the
// intervening if (CR 603.4) is read off the turn tally as the end step
// begins and again as the trigger resolves (revolt.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3d081701-514b-4611-b5a1-239b8f335fb6",
		Name:            "Insectoid Exterminator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfRevolt("Insectoid Exterminator — scry 1",
				func(g *game.Game, item *game.StackItem) error {
					return Scry{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
