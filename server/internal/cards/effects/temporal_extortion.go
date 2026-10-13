package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temporal Extortion — Sorcery {B}{B}{B}{B}:
//
//	"When you cast this spell, any player may pay half their life,
//	 rounded up. If a player does, counter Temporal Extortion.
//	 Take an extra turn after this one."
//
// ADR 0150 §3 (#1903): the payment is offered to every player in APNAP
// order from the active player, the caster included, and each player's
// price is half THEIR life, rounded up, worked out as they are asked
// (HalfTheirLifeRoundedUp). A player whose life total can't change
// can't pay (CR 119.8). The offer holds the spell on the stack until it
// is answered, and the first payer counters it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13da5255-2fc5-43a4-a44c-13a5d9ff77c3",
		Name:         "Temporal Extortion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			anyPlayerMayCounterThisSpell("Temporal Extortion", "pay half your life, rounded up", UpkeepPayment{}, HalfTheirLifeRoundedUp),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return TakeExtraTurn{}.Apply(ctx)
		},
	})
}
