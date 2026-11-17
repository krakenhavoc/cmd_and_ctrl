package effects

// Bloated Toad — Creature — Frog {2}{G}, 2/2 (Urza's Legacy, #5623):
//
//	"Protection from blue
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Disciple of Grace's shape with protection from blue: protection is the
// printed keyword (CR 702.16a) and cycling is the shared Cycling row
// (CR 702.29a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5ddcc7a5-c869-49fd-baa1-ab5f5f456f90",
		Name:            "Bloated Toad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from blue"},
		Activated:       []ActivatedAbility{Cycling("{2}")},
	})
}
