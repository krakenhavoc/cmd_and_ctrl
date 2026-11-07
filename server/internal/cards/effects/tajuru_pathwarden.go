package effects

// Tajuru Pathwarden — Creature — Elf Warrior Ally {4}{G}, 5/4
// (Foundations, #4465):
//
//	"Vigilance
//	 Trample"
//
// Two printed keywords the engine enforces from PrintedKeywords:
// attacking doesn't tap it, and its excess combat damage tramples over.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8c7cd66-d8bf-4a99-add5-1d5651d4e025",
		Name:            "Tajuru Pathwarden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
	})
}
