package effects

// Longbow Archer — Creature — Human Soldier Archer {W}{W}, 2/2 (Visions, #5631):
//
//	"Reach (This creature can block creatures with flying.)
//	 First strike"
//
// Both are engine keywords carried as printed data: reach lets it block
// a creature with flying (CR 702.17b), and first strike puts its combat
// damage in the first damage step (CR 702.7b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "51531db1-ba59-4ccc-9a91-54de2887fc8b",
		Name:            "Longbow Archer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "first strike"},
	})
}
