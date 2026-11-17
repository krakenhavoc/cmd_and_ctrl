package effects

// Blitz Automaton — Artifact Creature — Construct {7}, 6/4 (The
// Brothers' War, #5656): "Prototype {2}{R} — 3/2", "Haste".
//
// Prototype is the card's second face (CR 718, game/prototype.go): the
// deck importer builds it from the inset, and a cast of face 1 is the
// prototyped spell. Nothing here spells it out. Haste is a printed
// keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "19f344ee-031d-4346-9553-c5d32b02df43",
		Name:            "Blitz Automaton",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
