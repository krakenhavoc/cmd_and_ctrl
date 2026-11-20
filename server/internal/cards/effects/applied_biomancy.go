package effects

// Applied Biomancy — Instant {G}{U} (Ravnica Allegiance, #5521):
//
//	"Choose one or both —
//	 • Target creature gets +1/+1 until end of turn.
//	 • Return target creature to its owner's hand."
//
// Choose one or both (CR 700.2): Borrowed Malevolence's +1/+1 bullet and
// Mystic Confluence's bounce, each with its own target, run in printed
// order (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93ee2ea1-4f9b-47dd-b7d4-b0ed5a26d719",
		Name:         "Applied Biomancy",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeWithPurpose(ModeDoing("Target creature gets +1/+1 until end of turn.",
				TargetCreature("target creature"),
				BoostTheModesTarget(1, 1, "Applied Biomancy")), ForTargets(BenefitUntilEOT(0, 1, 1))),
			ModeDoing("Return target creature to its owner's hand.",
				TargetCreature("target creature"),
				BounceTheModesTarget),
		),
	})
}
