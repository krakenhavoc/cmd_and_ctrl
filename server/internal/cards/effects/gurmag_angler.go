package effects

// Gurmag Angler — Creature — Zombie Fish 5/5 {6}{B}:
//
//	"Delve"
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "315772f2-abd2-4681-b8c8-1db4b0ccbbcd",
		Name:         "Gurmag Angler",
		Completeness: CompletenessFull,
		Delve:        true,
	})
}
