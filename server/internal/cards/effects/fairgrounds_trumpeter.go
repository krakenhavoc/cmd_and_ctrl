package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fairgrounds Trumpeter — Creature — Elephant {2}{G}, 2/2:
//
//	"At the beginning of each end step, if a +1/+1 counter was put on a
//	 permanent under your control this turn, put a +1/+1 counter on this
//	 creature."
//
// An intervening "if" (CR 603.4): checked as the end step begins and again
// as the trigger resolves, off the turn's event log. Who put the counter
// is not asked; the permanent must be on the battlefield under your
// control, which counts counters a permanent entered with (CR 122.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3dfc6a52-d58c-4455-96fe-ec193743d67a",
		Name:         "Fairgrounds Trumpeter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(condPlusOneCounterPutOnYourPermanent{},
				On(game.EventStepBegan,
					AllOf(StepBegan(game.StepEnd, false),
						func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
							return g.PlusOneCounterPutOnYourPermanentThisTurnForEffect(source.Controller)
						}),
					"Fairgrounds Trumpeter — put a +1/+1 counter on this creature",
					Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
