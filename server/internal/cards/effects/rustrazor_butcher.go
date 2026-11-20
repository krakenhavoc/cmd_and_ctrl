package effects

// Rustrazor Butcher — Creature — Goblin Warrior {1}{R}, 1/2 (Shadowmoor,
// #5591):
//
//	"First strike
//	 Wither (This deals damage to creatures in the form of -1/-1
//	 counters.)"
//
// First strike is a printed keyword. Wither rides PrintedKeywords (ADR
// 0056, CR 702.80a): its damage to a creature is dealt as -1/-1 counters,
// as on Stigma Lasher.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58fd5a3c-83ca-4f6a-99db-21d5e03f2345",
		Name:            "Rustrazor Butcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "wither"},
	})
}
