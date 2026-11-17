package effects

// Goblin Berserker — Creature — Goblin Berserker {3}{R}, 2/2 (Urza's
// Destiny, #5621):
//
//	"First strike, haste"
//
// Two printed keywords: the combat damage step reads first strike, and
// the attack check lets it attack the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5f71f205-9551-416a-bba6-55bc65d11c85",
		Name:            "Goblin Berserker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
	})
}
