package effects

// Zealous Strike — {1}{W} Instant (Avacyn Restored, #5570):
//
//	"Target creature gets +2/+2 and gains first strike until end of
//	 turn."
//
// Interjection's text, the shared socB27TargetGetsAndGains body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d820ddf4-315d-470a-8d66-663a76a240c6",
		Name:         "Zealous Strike",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "first strike")),
		OnResolve:    socB27TargetGetsAndGains(2, 2, "first strike", "Zealous Strike"),
	})
}
