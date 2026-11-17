package effects

// Phyrexian Debaser — Creature — Phyrexian Carrier {3}{B}, 2/2 (Urza's
// Legacy, #5623):
//
//	"Flying
//	 {T}, Sacrifice this creature: Target creature gets -2/-2 until end
//	 of turn."
//
// A Carrier through the shared ulgB1SacrificeToShrink row at -2/-2: {T} and
// sacrificing it are both costs, so it waits out summoning sickness
// (CR 302.6) and is gone before the ability resolves. Flying is a printed
// keyword (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "554de904-086c-428c-9f32-b0b2498a6fd5",
		Name:            "Phyrexian Debaser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			ulgB1SacrificeToShrink("Phyrexian Debaser", 2),
		},
	})
}
