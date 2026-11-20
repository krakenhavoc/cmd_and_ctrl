package effects

// Tattered Drake — Creature — Zombie Drake {4}{U}, 2/2 (Ravnica: City of
// Guilds, #5600):
//
//	"Flying
//	 {B}: Regenerate this creature."
//
// Flying is printed data; the regeneration row is regenerateThisRow (CR
// 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1199a252-3c7c-4ae4-b063-dc132bfbea2f",
		Name:            "Tattered Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{regenerateThisRow("{B}")},
	})
}
