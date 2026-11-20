package effects

// Grazing Whiptail — Creature — Dinosaur {2}{G}{G}, 3/4 (Ixalan, #5532):
//
//	"Reach"
//
// Reach is an engine keyword carried as printed data: it can block
// creatures with flying (CR 702.17b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "94399b1f-e66a-422e-ae41-6013fe741e8c",
		Name:            "Grazing Whiptail",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
