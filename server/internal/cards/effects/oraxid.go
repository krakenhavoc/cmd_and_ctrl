package effects

// Oraxid — Creature — Crab Beast {3}{U}, 2/3 (Nemesis, #5618):
//
//	"Protection from red"
//
// One printed keyword, enforced by the engine: a red creature can't
// block it (CR 702.16f), red spells and abilities can't target it
// (CR 702.16b), and damage from red sources is prevented (CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7d086ae5-738c-4685-80ae-66bd9379d8f7",
		Name:            "Oraxid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from red"},
	})
}
