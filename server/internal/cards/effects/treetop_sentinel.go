package effects

// Treetop Sentinel — Creature — Bird Soldier {2}{U}{U}, 2/3 (Odyssey, #5613):
//
//	"Flying, protection from green"
//
// Aven Smokeweaver's body with protection from green: flying (CR
// 702.9b), and green spells, sources and blockers are refused (CR
// 702.16b, 702.16e, 702.16f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "786afd52-0eae-4453-b1bd-cffc21699849",
		Name:            "Treetop Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from green"},
	})
}
