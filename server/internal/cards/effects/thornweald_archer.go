package effects

// Thornweald Archer — Creature — Elf Archer {1}{G}, 2/1 (Foundations,
// #4465):
//
//	"Reach
//	 Deathtouch"
//
// Two printed keywords the engine enforces from PrintedKeywords: it can
// block a flier, and any damage it deals to a creature is lethal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5ba87f5-3589-4a8b-ad65-2ae2523e98e8",
		Name:            "Thornweald Archer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
	})
}
