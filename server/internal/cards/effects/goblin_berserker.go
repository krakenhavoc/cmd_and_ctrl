package effects

// Goblin Berserker — Creature — Goblin Berserker {3}{R}, 2/2 (Urza's Destiny, #5621):
//
//	"First strike, haste"
//
// Goblin Striker's two keywords on a 2/2. Haste lets it attack and use
// {T} the turn it arrives (CR 702.10a); first strike is read by the
// combat damage step.
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
