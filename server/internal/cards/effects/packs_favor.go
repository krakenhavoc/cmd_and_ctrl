package effects

// Pack's Favor — Instant {2}{G} (Guilds of Ravnica, #5525):
//
//	"Convoke
//	 Target creature gets +3/+3 until end of turn."
//
// Convoke is the shared tap cost (CR 702.51a); the pump is Giant Growth's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "072288da-5ba0-4176-b0e6-b3e43ecab2ec",
		Name:         "Pack's Favor",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3)),
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(3, 3, "Pack's Favor — +3/+3 until end of turn"),
	})
}
