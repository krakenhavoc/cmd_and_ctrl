package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Michelangelo, Game Master — {2}{G} Legendary Creature — Mutant Ninja
// Turtle 3/3 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Disappear — At the beginning of your end step, if a permanent left
//	 the battlefield under your control this turn, put a +1/+1 counter on
//	 Michelangelo."
//
// Disappear is an ability word (CR 207.2c) over revolt's condition: the
// end-step intervening if (CR 603.4) of revolt.go. A Michelangelo that has
// left by the time the trigger resolves gets nothing (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "594c6b1f-ff0e-4122-8b29-852632a22a79",
		Name:         "Michelangelo, Game Master",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourEndStepIfRevolt("Michelangelo, Game Master — put a +1/+1 counter on Michelangelo",
				putACounterOnThis(game.CounterPlusOne)),
		},
	})
}
