package effects

// Juvenile Gloomwidow — Creature — Spider {G}{G}, 1/3 (Shadowmoor, #5591):
//
//	"Reach (This creature can block creatures with flying.)
//	 Wither (This deals damage to creatures in the form of -1/-1
//	 counters.)"
//
// Reach is a printed keyword. Wither rides PrintedKeywords (ADR 0056, CR
// 702.80a): its damage to a creature is dealt as -1/-1 counters, as on
// Stigma Lasher.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d2546d25-07cd-4830-ba66-34bdbf8a245c",
		Name:            "Juvenile Gloomwidow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "wither"},
	})
}
