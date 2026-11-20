package effects

// Lost Leonin — Creature — Phyrexian Cat Soldier {1}{W}, 2/1 (New Phyrexia, #5575):
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)"
//
// Infect is the engine's damage tail (CR 702.90).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c21b5c5b-9191-4748-b111-a51631e0bc5a",
		Name:            "Lost Leonin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
	})
}
