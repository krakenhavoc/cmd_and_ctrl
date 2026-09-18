package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pact of Negation — Instant {0}:
//
//	"Counter target spell.
//	 At the beginning of your next upkeep, pay {3}{U}{U}. If you
//	 don't, you lose the game."
//
// The card in the S28 "free-cast" group that needs NO alternative
// cost at all, which is worth saying out loud because the issue
// listed it next to Force of Will. Pact's printed mana cost really is
// {0} — there is nothing to replace. What it has instead is a debt,
// and the debt is an ordinary CR 603.7 delayed triggered ability the
// engine has had since S22.
//
// Three details the delayed trigger gets right for free:
//
//   - "YOUR next upkeep" — ControllerTurnOnly, so an opponent's
//     upkeep in between does not collect. (Arcane Denial's "the next
//     turn's upkeep" is the other spelling and leaves the flag off.)
//   - It uses the stack, so the payment prompt does not happen in a
//     vacuum: everyone gets a response window first.
//   - It survives the Pact itself — by upkeep the card is in a
//     graveyard, and nothing about the debt is stored on it.
//
// The payment reuses the S19 pay-unless prompt with the consequence
// on the decline branch, and a "pay" the player cannot fund degrades
// to a decline server-side, which is exactly the printed outcome: if
// you don't pay, you lose.
//
// Losing goes through LoseTheGameForEffect, which today borrows the
// AttemptedEmptyDraw flag so the loss lands at the next state-based
// check. ADR 0057 sub-PR 2 makes an effect loss immediate (CR 104.3e)
// and rewrites this comment.
func init() {
	Register(Spec{
		OracleID:     "f3e213a4-ba5a-468a-93b3-c0a34e1bd725",
		Name:         "Pact of Negation",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) > 0 {
				if err := (CounterTarget{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
			}
			// The debt is owed whether or not the counter did
			// anything — a Pact whose target left the stack still
			// costs you {3}{U}{U} next upkeep.
			return ScheduleDelayedTrigger{
				At:                 game.StepUpkeep,
				ControllerTurnOnly: true,
				Label:              "Pact of Negation — pay {3}{U}{U} or lose the game",
				Effect:             pactPayment("{3}{U}{U}", "Pact of Negation"),
			}.Apply(ctx)
		},
	})
}

// pactPayment builds the Pact cycle's upkeep trigger: prompt the
// controller for `cost`, and on a decline — or a "pay" they cannot
// fund — they lose the game.
//
// `cost` and `name` are copied strings, so the closure captures no
// game state and survives Clone / undo, the same contract
// manaDrainRefund follows.
func pactPayment(cost, name string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		payer := item.Controller
		return PayUnless{
			Chooser:  payer,
			Cost:     cost,
			Question: name + " — pay " + cost + " or lose the game",
			OnDecline: func(ctx *Context) error {
				return ctx.Game.LoseTheGameForEffect(payer)
			},
		}.Apply(NewContext(g, item))
	}
}
