package effects

// Goliath Beetle — Creature — Insect {2}{G}, 3/1 (Urza's Destiny, #5621):
//
//	"Trample"
//
// A vanilla 3/1 with trample, which the combat damage step reads when
// it assigns excess damage past its blockers (CR 702.19b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "86ab4400-fbdd-4c18-a893-441286a9d7d0",
		Name:            "Goliath Beetle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
