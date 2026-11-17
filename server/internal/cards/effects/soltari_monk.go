package effects

// Soltari Monk — Creature — Soltari Monk Cleric {W}{W}, 2/1
// (Tempest, #5628):
//
//	"Protection from black
//	 Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)"
//
// Two printed keywords, both enforced by the engine. With shadow, only a
// creature with shadow can block it (CR 702.28b); with protection from
// black, a black one can't (CR 702.16f), so its only blockers are
// nonblack creatures with shadow. Black spells and abilities can't
// target it (CR 702.16b), and damage from black sources is prevented
// (CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5fed5377-07cf-4ad9-86fd-1efe1f182088",
		Name:            "Soltari Monk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from black", "shadow"},
	})
}
