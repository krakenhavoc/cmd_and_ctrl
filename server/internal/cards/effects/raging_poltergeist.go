package effects

// Raging Poltergeist — {4}{R} Creature — Spirit 6/1 (Avacyn Restored, #5570):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// grizzly_bears.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "15d460da-9110-4f33-a24f-8582a2bcd96f",
		Name:         "Raging Poltergeist",
		Completeness: CompletenessFull,
	})
}
