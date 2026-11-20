package effects

// Living Tempest — Creature — Elemental {4}{U}, 3/3 (Zendikar Rising,
// #5508):
//
//	"Flash
//	 Flying"
//
// Both are engine keywords carried as printed data: flash (CR 702.8a)
// and flying (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d684553a-f225-49af-90eb-399f55ef65df",
		Name:            "Living Tempest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
