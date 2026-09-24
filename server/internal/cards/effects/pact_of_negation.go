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
// Losing goes through LoseTheGame, and it is IMMEDIATE (CR 104.3e —
// an effect loss has no state-based-action clause; ADR 0057
// Decision 3): the player leaves the moment they decline, and a
// "can't lose the game" gate (Platinum Angel) read at that moment
// stops it. The loser is the active player — the debt is collected
// in their own upkeep — so the turn moves on at the next state-based
// check rather than inside the answer.
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
				Body:               pactPaymentBody,
				Params:             game.EffectParams{Cost: "{3}{U}{U}", Name: "Pact of Negation"},
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
//
// UpkeepPayUnless, not PayUnless: the debt is the active player's
// own, owed in their own upkeep, and what hangs on the answer is the
// largest consequence in the game. The table does not leave the
// upkeep until it is answered (#997, CR 500.4). It used PayUnless
// until then, and a table could take the turn — and the Pact's
// controller could untap, draw and attack — with the {3}{U}{U} still
// unpaid and the loss still pending.
func pactPayment(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	payer := item.Controller
	return UpkeepPayUnless{
		Chooser:  payer,
		Cost:     p.Cost,
		Question: p.Name + " — pay " + p.Cost + " or lose the game",
		OnDecline: func(ctx *Context) error {
			return LoseTheGame{Player: payer}.Apply(ctx)
		},
	}.Apply(NewContext(g, item))
}
