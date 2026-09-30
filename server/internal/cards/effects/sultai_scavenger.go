package effects

// Sultai Scavenger — Creature — Bird Warrior 3/3 {5}{B}:
//
//	"Delve. Flying"
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "837f72a9-9522-46e9-a48c-c6ff6d0bbe67",
		Name:            "Sultai Scavenger",
		Completeness:    CompletenessFull,
		Delve:           true,
		PrintedKeywords: []string{"flying"},
	})
}
