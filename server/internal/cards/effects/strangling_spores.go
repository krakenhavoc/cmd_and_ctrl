package effects

// Strangling Spores — Instant {3}{B} (Core Set 2019, #5527):
//
//	"Target creature gets -3/-3 until end of turn."
//
// A -3/-3 until end of turn on one target, through the shared
// boostTheTargetUntilEOT.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "444493a6-0f90-4847-9619-4ea8fbf7cf6f",
		Name:         "Strangling Spores",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, -3, -3)),
		OnResolve:    boostTheTargetUntilEOT(-3, -3, "Strangling Spores — -3/-3 until end of turn"),
	})
}
