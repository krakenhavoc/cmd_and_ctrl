package effects

// Elvish Lookout — Creature — Elf {G}, 1/1 (Urza's Destiny, #5621):
//
//	"Shroud (This creature can't be the target of spells or abilities.)"
//
// Deadly Insect's shroud on a 1/1 (CR 702.18a): no spell or ability
// can target it, its controller's included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e9aea99-1eee-4bb6-9043-a1ac2884d071",
		Name:            "Elvish Lookout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
	})
}
