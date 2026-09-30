package effects

// Shambling Attendants — Creature — Zombie 3/5 {7}{B}:
//
//	"Delve. Deathtouch"
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cd17bd00-90b1-4b89-9368-2d8735159e47",
		Name:            "Shambling Attendants",
		Completeness:    CompletenessFull,
		Delve:           true,
		PrintedKeywords: []string{"deathtouch"},
	})
}
