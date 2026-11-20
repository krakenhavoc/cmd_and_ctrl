package effects

// Vampire Champion — Creature — Vampire Soldier {3}{B}, 3/3 (Rivals of Ixalan, #5531):
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough
//	 to destroy it.)"
//
// Deathtouch is an engine keyword carried as printed data: any damage it
// deals to a creature is lethal (CR 702.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a48d2be2-395a-4377-a2ad-51dbbc5b4e17",
		Name:            "Vampire Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
