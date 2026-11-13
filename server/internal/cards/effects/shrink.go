package effects

// Shrink — Instant {G} (Homelands, #5635):
//
//	"Target creature gets -5/-0 until end of turn."
//
// A layer 7c change (CR 613.4c) that ends at cleanup (CR 514.2). Power
// can go below zero, and a creature with 0 or less power assigns no
// combat damage (CR 510.1a). If the target has left by the time Shrink resolves, it does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "252330b1-67cf-4d9e-a413-917ba61e731f",
		Name:         "Shrink",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(HarmUntilEOT(0, -5, 0)),
		OnResolve:    boostTheTargetUntilEOT(-5, 0, "Shrink — -5/-0 until end of turn"),
	})
}
