package effects

// Goblin Chariot — Creature — Goblin Warrior {2}{R}, 2/2
// (Starter 1999, #5620):
//
//	"Haste (This creature can attack and {T} as soon as it comes under your control.)"
//
// Raging Goblin's keyword: haste is printed data, so it can attack
// and tap the turn it arrives (CR 702.10b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6ef0dd51-31df-4fd3-8153-49a427a7fdcf",
		Name:            "Goblin Chariot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
