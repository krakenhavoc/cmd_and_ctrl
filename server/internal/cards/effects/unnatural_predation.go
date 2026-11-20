package effects

// Unnatural Predation — Instant {G} (Mirrodin Besieged,
// #5576):
//
//	"Target creature gets +1/+1 and gains trample until end of turn."
//
// Fanatical Strength's body (socB27TargetGetsAndGains) at +1/+1: both
// halves land on the target if it is still legal (CR 608.2b) and end at
// cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "120446cc-f55b-40db-915d-2669c1083e56",
		Name:         "Unnatural Predation",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1, "trample")),
		OnResolve:    socB27TargetGetsAndGains(1, 1, "trample", "Unnatural Predation"),
	})
}
