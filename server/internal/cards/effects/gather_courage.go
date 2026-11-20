package effects

// Gather Courage — Instant {G} (Ravnica: City of Guilds, #5600):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Target creature gets +2/+2 until end of turn."
//
// Convoke is the engine's tap cost (CR 702.51); the pump is the one target's
// +2/+2 until cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae033eb6-47ec-4b9e-9cdb-6d023aca07d1",
		Name:         "Gather Courage",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		OnResolve:    boostTheTargetUntilEOT(2, 2, "Gather Courage — +2/+2"),
	})
}
