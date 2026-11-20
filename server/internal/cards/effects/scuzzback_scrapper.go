package effects

// Scuzzback Scrapper — Creature — Goblin Warrior {R/G}, 1/1 (Shadowmoor,
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
		OracleID:        "508c673c-2b11-4b7d-8f4b-ff3a364eb4fe",
		Name:            "Scuzzback Scrapper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"wither"},
	})
}
