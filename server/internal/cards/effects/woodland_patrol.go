package effects

// Woodland Patrol — Creature — Human Scout {2}{G}, 3/2 (Eldritch Moon, #5541):
//
//	"Vigilance"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ce31cfc4-a109-4cf5-b8f9-210267a98114",
		Name:            "Woodland Patrol",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
