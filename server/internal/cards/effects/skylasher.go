package effects

// Skylasher — Creature — Insect {1}{G}, 2/2:
//
//	"Flash
//	 This spell can't be countered.
//	 Reach, protection from blue"
//
// Flash (CR 702.8), reach and protection from blue (CR 702.16) are printed
// keywords. "This spell can't be countered" is Spec.CantBeCountered,
// Toski's field.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "153a4c11-e75f-46d7-b684-17d500ea13e6",
		Name:            "Skylasher",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"flash", "reach", "protection from blue"},
	})
}
