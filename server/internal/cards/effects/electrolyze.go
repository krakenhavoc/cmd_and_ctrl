package effects

// Electrolyze — Instant {1}{U}{R} (#1658, unblocked by #1656):
//
//	"Electrolyze deals 2 damage divided as you choose among one or
//	 two targets.
//	 Draw a card."
//
// Forked Bolt's split with card advantage bolted on. The draw is
// unconditional and not part of the division — it happens whether or
// not either target is still legal at resolution (CR 608.2b only
// touches the divided damage).
func init() {
	Register(Spec{
		OracleID:     "07b222d7-24f2-4994-9004-ff6672ebe161",
		Name:         "Electrolyze",
		Completeness: CompletenessFull,
		Targets:      TargetAny().WithCount(1, 2).Dividing(Divide(2)),
		// Two steps, so the draw waits for a prompt the damage asks
		// (Khârn the Betrayer's question, a CR 616 order), CR 608.2c
		// (#4019).
		OnResolve: ResolveSteps(Instruction(DealDividedDamage), DrawCards{N: 1}),
	})
}
