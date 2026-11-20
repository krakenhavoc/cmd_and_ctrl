package effects

// Kessig Recluse — Creature — Spider {2}{G}{G}, 2/3 (Dark Ascension, #5571):
//
//	"Reach (This creature can block creatures with flying.)
//	 Deathtouch (Any amount of damage this deals to a creature is
//	 enough to destroy it.)"
//
// Reach and deathtouch are printed keywords, as on Thornweald Archer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "61c3b999-4f5d-4018-a9ee-ef894fd1ef6c",
		Name:            "Kessig Recluse",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
	})
}
