package effects

// Tombstalker — Creature — Demon 7/7 {6}{B}{B}:
//
//	"Flying. Delve"
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "944735ad-3f17-421a-b8fa-40f3bb8456f2",
		Name:            "Tombstalker",
		Completeness:    CompletenessFull,
		Delve:           true,
		PrintedKeywords: []string{"flying"},
	})
}
