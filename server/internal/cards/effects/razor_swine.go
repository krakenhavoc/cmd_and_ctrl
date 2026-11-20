package effects

// Razor Swine — Creature — Phyrexian Boar {2}{R}, 2/1 (New Phyrexia, #5575):
//
//	"First strike
//	 Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)"
//
// Both are engine keywords (CR 702.7, CR 702.90).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4d5649a1-5716-4560-a956-84300bf6a2f4",
		Name:            "Razor Swine",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "infect"},
	})
}
