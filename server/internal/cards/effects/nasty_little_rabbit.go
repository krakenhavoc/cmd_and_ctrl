package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nasty Little Rabbit — Creature — Rabbit {G}, 1/2 (The Hobbit):
//
//	"Ferocious — At the beginning of combat on your turn, if you
//	 control a creature with power 4 or greater, put a +1/+1 counter on
//	 this creature."
//
// An intervening "if" (CR 603.4): checked as combat begins to put the
// trigger on the stack, and again as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee86cce6-c7c1-40a6-896b-cde9b86bb532",
		Name:         "Nasty Little Rabbit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
					b24ControlsCreatureWithPowerAtLeast(g, source.Controller, 4)
			}, "Nasty Little Rabbit — ferocious: a +1/+1 counter on this creature",
				func(g *game.Game, item *game.StackItem) error {
					if !b24ControlsCreatureWithPowerAtLeast(g, item.Controller, 4) {
						return nil
					}
					return Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})(g, item)
				}),
		},
	})
}
