package effects

// Savannah Lions — Creature — Cat {W}, 2/1 (Foundations, #4465):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// fire_elemental.go). The cost, types and body ride game.Card from
// Scryfall, so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60ba93eb-39e6-4af2-9c66-cd38f72daff2",
		Name:         "Savannah Lions",
		Completeness: CompletenessFull,
	})
}
