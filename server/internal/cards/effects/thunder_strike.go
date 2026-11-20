package effects

// Thunder Strike — Instant {1}{R} (Magic 2011, #5579):
//
//	"Target creature gets +2/+0 and gains first strike until end of
//	 turn."
//
// Fit of Rage's pump and grant at +2/+0: a layer 7c change and a layer 6
// grant that both end at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5f25e80-7a6a-4629-8201-901e0fcd4a02",
		Name:         "Thunder Strike",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0, "first strike")),
		OnResolve:    socB27TargetGetsAndGains(2, 0, "first strike", "Thunder Strike"),
	})
}
