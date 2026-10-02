package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Intervention Pact — Instant {0}:
//
//	"The next time a source of your choice would deal damage to you this turn, prevent that damage. You gain life equal to the damage prevented this way.
//	 At the beginning of your next upkeep, pay {1}{W}{W}. If you don't, you lose the game."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8, 609.7a), with
// "You gain life equal to the damage prevented this way" as its
// additional effect (CR 615.5, owner decision 4). The pact's upkeep
// payment is Pact of Negation's delayed trigger (pactPaymentBody).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9f9882f5-2338-4881-b418-b15348a462b4",
		Name:         "Intervention Pact",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (ScheduleDelayedTrigger{
				At:                 game.StepUpkeep,
				ControllerTurnOnly: true,
				Label:              "Intervention Pact — pay {1}{W}{W} or lose the game",
				Body:               pactPaymentBody,
				Params:             game.EffectParams{Cost: "{1}{W}{W}", Name: "Intervention Pact"},
			}).Apply(ctx); err != nil {
				return err
			}
			return PreventNextDamageFromChosenSource(ShieldYou).WithThen(preventedGainLifeBody).Apply(ctx)
		},
	})
}
