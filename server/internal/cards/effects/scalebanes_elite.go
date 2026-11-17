package effects

// Scalebane's Elite — Creature — Human Soldier {3}{G}{W}, 4/4 (Visions, #5631):
//
//	"Protection from black"
//
// An engine keyword carried as printed data: black spells and abilities
// from black sources can't target it, and damage from black sources is
// prevented (CR 702.16b, CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c04dc6fe-06fc-4e5b-87da-6b8b262a214f",
		Name:            "Scalebane's Elite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
	})
}
