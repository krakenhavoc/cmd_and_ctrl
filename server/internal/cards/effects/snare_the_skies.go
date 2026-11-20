package effects

// Snare the Skies — {G} Instant (Avacyn Restored, #5570):
//
//	"Target creature gets +1/+1 and gains reach until end of turn. (It
//	 can block creatures with flying.)"
//
// The shared socB27TargetGetsAndGains body at +1/+1 and reach.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d011e22-335b-465f-9111-e95a72ebffe9",
		Name:         "Snare the Skies",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "reach")),
		OnResolve:    socB27TargetGetsAndGains(1, 1, "reach", "Snare the Skies"),
	})
}
