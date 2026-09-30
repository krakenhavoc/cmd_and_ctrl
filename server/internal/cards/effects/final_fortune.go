package effects

// Final Fortune — Instant {R}{R}:
//
//	"Take an extra turn after this one. At the beginning of that turn's
//	 end step, you lose the game."
//
// TakeExtraTurnThenLose (extra_turns.go): the turn is queued (CR 500.7)
// and the loss is a delayed trigger bound to THAT turn
// (DelayedTrigger.OnExtraTurn, ADR 0059 Decision 8). So:
//
//   - it does not fire in the end step of the turn Final Fortune was
//     cast in — an instant can be cast in an end step, and an unbound
//     "next end step" trigger would pick the wrong one;
//   - it uses the stack, so it can be answered;
//   - if the extra turn never begins, or ends before its end step, the
//     trigger is swept and the player does not lose (2004-10-04
//     ruling). A second extra turn cast during this one is taken AFTER
//     it (CR 500.7), so it does not rescue the player.
//
// The loss goes through ADR 0057's LoseTheGame, so a "you can't lose
// the game" effect (Platinum Angel) still stops it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8d0adc5c-3fdd-4e22-b783-5651f8e57b65",
		Name:         "Final Fortune",
		Completeness: CompletenessFull,
		OnResolve:    extraTurnThenLoseAtItsEndStep("Final Fortune — you lose the game"),
	})
}
