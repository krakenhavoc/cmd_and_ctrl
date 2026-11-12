package effects

// Breach — Instant {2}{B} (Urza's Saga, #5624):
//
//	"Target creature gets +2/+0 and gains fear until end of turn. (It
//	 can't be blocked except by artifact creatures and/or black
//	 creatures.)"
//
// Fanatical Strength's body (socB27TargetGetsAndGains) with fear: both
// halves land on the target if it is still legal (CR 608.2b) and end at
// cleanup. Fear is the canonical keyword the block check honours
// (CR 702.36a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c412ae1-3499-41fa-b3c4-70131c9bf0bd",
		Name:         "Breach",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 0, "fear")),
		OnResolve:    socB27TargetGetsAndGains(2, 0, "fear", "Breach"),
	})
}
