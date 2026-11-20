package effects

// Hydrosurge — Instant {U} (Magic 2013, #5568):
//
//	"Target creature gets -5/-0 until end of turn."
//
// Shrink in blue: a layer 7c change (CR 613.4c) that ends at cleanup
// (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c2059d5-9a2e-4b89-a617-9125eb0cd570",
		Name:         "Hydrosurge",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, -5, 0)),
		OnResolve:    boostTheTargetUntilEOT(-5, 0, "Hydrosurge — -5/-0 until end of turn"),
	})
}
