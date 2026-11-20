package effects

// Aegis of the Heavens — Instant {1}{W} (Core Set 2019, #5527):
//
//	"Target creature gets +1/+7 until end of turn."
//
// A +1/+7 until end of turn on one target, through the shared
// boostTheTargetUntilEOT.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8e1b38a-2a77-4cf1-8bef-31acb61ede10",
		Name:         "Aegis of the Heavens",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 7)),
		OnResolve:    boostTheTargetUntilEOT(1, 7, "Aegis of the Heavens — +1/+7"),
	})
}
