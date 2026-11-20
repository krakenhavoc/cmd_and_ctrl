package effects

// Bull Rush — Instant {R} (Worldwake, #5582):
//
//	"Target creature gets +2/+0 until end of turn."
//
// Rouse's pump, without its alternative cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0140cf53-71ef-4d37-9197-48aab35ca4a2",
		Name:         "Bull Rush",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(2, 0, "Bull Rush — +2/+0"),
	})
}
