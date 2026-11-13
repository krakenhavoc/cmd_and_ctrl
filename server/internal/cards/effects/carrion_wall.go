package effects

// Carrion Wall — Creature — Wall {1}{B}{B}, 3/2
// (Nemesis, #5618):
//
//	"Defender (This creature can't attack.)
//	 {1}{B}: Regenerate this creature."
//
// Defender is printed data (CR 702.3b). The second row is the shared
// regenerateThisRow: one regeneration shield on the Wall (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ce8750d4-1b74-4b02-afd8-6b639c1a9a78",
		Name:            "Carrion Wall",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated:       []ActivatedAbility{regenerateThisRow("{1}{B}")},
	})
}
