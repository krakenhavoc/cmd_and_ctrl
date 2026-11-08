package effects

// Skyraker Giant — Creature — Giant {2}{R}{R}, 4/3 (Foundations, #4465):
//
//	"Reach"
//
// One printed keyword the engine enforces from PrintedKeywords: it can
// block a creature with flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d481630c-83c4-47d2-a770-db331154d2b9",
		Name:            "Skyraker Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
