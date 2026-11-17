package effects

// Rust Goliath — Artifact Creature — Construct {10}, 10/10 (The Brothers'
// War, #5656): "Prototype {3}{G}{G} — 3/5", "Reach, trample".
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// Reach and trample are printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61df991d-32e7-4b6b-b8e6-700ba915c3f2",
		Name:            "Rust Goliath",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
	})
}
