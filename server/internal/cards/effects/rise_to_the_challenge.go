package effects

// Rise to the Challenge — Instant {1}{R} (Born of the Gods,
// #5559):
//
//	"Target creature gets +2/+0 and gains first strike until end of turn."
//
// Interjection's body (socB27TargetGetsAndGains): both halves on the
// target still legal as it resolves, ending at cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7f3095fb-ffc0-48ed-8792-4157897a9ce7",
		Name:         "Rise to the Challenge",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0, "first strike")),
		Targets:      TargetCreature("target creature"),
		OnResolve:    socB27TargetGetsAndGains(2, 0, "first strike", "Rise to the Challenge"),
	})
}
