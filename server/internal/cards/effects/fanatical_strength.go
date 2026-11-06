package effects

// Fanatical Strength — Instant {1}{G} (Murders at Karlov Manor, #4469):
//
//	"Target creature gets +3/+3 and gains trample until end of turn."
//
// Staggering Size's body (socB27TargetGetsAndGains): both halves land on
// the target if it is still legal (CR 608.2b) and end at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b92e97f5-a307-471f-be06-afa4fcdab757",
		Name:         "Fanatical Strength",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "trample")),
		OnResolve:    socB27TargetGetsAndGains(3, 3, "trample", "Fanatical Strength"),
	})
}
