package effects

// Malach of the Dawn — Creature — Angel {2}{W}{W}, 2/4 (Planar Chaos, #5595):
//
//	"Flying
//	 {W}{W}{W}: Regenerate this creature."
//
// Flying is printed data (CR 702.9b); regeneration is on the ability's own
// source (CR 701.19), lrwB1RegenerateThisRow.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c351a325-42cb-43ad-8648-c204e2317b50",
		Name:            "Malach of the Dawn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{lrwB1RegenerateThisRow("{W}{W}{W}")},
	})
}
