package effects

// Sickle Ripper — Creature — Elemental Warrior {1}{B}, 2/1 (Shadowmoor,
// #5591):
//
//	"Wither (This deals damage to creatures in the form of -1/-1
//	 counters.)"
//
// Wither rides PrintedKeywords (ADR 0056, CR 702.80a): its damage to a
// creature is dealt as -1/-1 counters, as on Stigma Lasher.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d097b223-c558-4592-abb9-44dc8eb52a61",
		Name:            "Sickle Ripper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"wither"},
	})
}
