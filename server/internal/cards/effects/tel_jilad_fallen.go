package effects

// Tel-Jilad Fallen — Creature — Phyrexian Elf Warrior {2}{G}{G}, 3/1 (Scars of Mirrodin, #5577)
//
//	"Protection from artifacts
//	 Infect (This creature deals damage to creatures in the form of -1/-1
//	 counters and to players in the form of poison counters.)"
//
// A keyword body: protection from artifacts (CR 702.16a) and infect
// (CR 702.90) are the engine's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e02740e6-3eba-4f82-831b-35a9f122f07f",
		Name:            "Tel-Jilad Fallen",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from artifacts", "infect"},
	})
}
