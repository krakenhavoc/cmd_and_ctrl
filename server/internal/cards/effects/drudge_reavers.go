package effects

// Drudge Reavers — Creature — Skeleton {3}{B}, 2/2 (Time Spiral, #5596):
//
//	"Flash
//	 {B}: Regenerate this creature."
//
// Flash is printed; the regeneration is the shared regenerateThisRow.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3edc147f-a385-413a-99b6-27d2cf8cf930",
		Name:            "Drudge Reavers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated:       []ActivatedAbility{regenerateThisRow("{B}")},
	})
}
