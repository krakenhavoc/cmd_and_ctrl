package effects

// Quilled Slagwurm — Creature — Phyrexian Wurm {4}{G}{G}{G}, 8/8 (Mirrodin Besieged,
// #5576):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see
// ordinary_bear.go). The cost, types and body ride game.Card from Scryfall,
// so the Spec is the key and the name.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5114f3e0-339d-46d6-9d44-97fa0afbd7c5",
		Name:         "Quilled Slagwurm",
		Completeness: CompletenessFull,
	})
}
