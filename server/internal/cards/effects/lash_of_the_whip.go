package effects

// Lash of the Whip — Instant {4}{B} (Theros, #5561):
//
//	"Target creature gets -4/-4 until end of turn."
//
// The shared until-end-of-turn shrink of the one target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2ac369a-f6d4-4b58-8695-a7c5b906e999",
		Name:         "Lash of the Whip",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -4, -4)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(-4, -4, "Lash of the Whip — -4/-4 until end of turn"),
	})
}
