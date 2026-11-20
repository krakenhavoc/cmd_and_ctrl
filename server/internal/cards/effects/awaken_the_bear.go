package effects

// Awaken the Bear — Instant {2}{G} (Khans of Tarkir, #5554):
//
//	"Target creature gets +3/+3 and gains trample until end of turn."
//
// Staggering Size's body at +3/+3 (socB27TargetGetsAndGains); both
// halves end in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53dc9795-d26a-4821-afad-8d77d279ffe3",
		Name:         "Awaken the Bear",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "trample")),
		OnResolve:    socB27TargetGetsAndGains(3, 3, "trample", "Awaken the Bear"),
	})
}
