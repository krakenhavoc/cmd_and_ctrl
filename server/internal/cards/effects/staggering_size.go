package effects

// Staggering Size — Instant {1}{G} (The Lost Caverns of Ixalan, #4470):
//
//	"Target creature gets +3/+3 and gains trample until end of turn."
//
// Interjection's body with +3/+3 and trample (socB27TargetGetsAndGains):
// a pump and a keyword grant locked to the target at resolution (CR
// 611.2c) that end in the cleanup step (CR 514.2). A target that is no
// longer legal gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "57bda813-ad4b-40ab-aa35-3f8343ff8668",
		Name:         "Staggering Size",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "trample")),
		OnResolve:    socB27TargetGetsAndGains(3, 3, "trample", "Staggering Size"),
	})
}
