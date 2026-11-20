package effects

// Dark Remedy — Instant {1}{B} (Core Set 2020, #5518):
//
//	"Target creature gets +1/+3 until end of turn."
//
// Artful Maneuver's pump body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7cc744b-2fbd-40e6-8a59-392535c17454",
		Name:         "Dark Remedy",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 3)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(1, 3, "Dark Remedy — +1/+3"),
	})
}
