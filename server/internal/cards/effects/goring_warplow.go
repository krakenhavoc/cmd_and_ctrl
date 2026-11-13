package effects

// Goring Warplow — Artifact Creature — Construct {6}, 5/4 (The Brothers'
// War, #5656): "Prototype {1}{B} — 1/1", "Deathtouch".
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// Deathtouch is a printed keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "02457c7f-eced-4bf3-9b61-309b4ceab2d5",
		Name:            "Goring Warplow",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
