package effects

// Enemy of the Guildpact — Creature — Spirit {4}{B}, 4/2 (Dissension, #5598):
//
//	"Protection from multicolored"
//
// Stonecoil Serpent's protection from multicolored, a whole-phrase quality
// of the protection grammar (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e5cf7b75-b59b-4cac-bfd2-13e452457d74",
		Name:            "Enemy of the Guildpact",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from multicolored"},
	})
}
