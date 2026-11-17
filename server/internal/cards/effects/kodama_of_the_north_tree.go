package effects

// Kodama of the North Tree — Legendary Creature — Spirit {2}{G}{G}{G},
// 6/4 (Champions of Kamigawa, #5603):
//
//	"Trample
//	 Shroud"
//
// Two printed keywords (CR 702.19, CR 702.18a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "15d8f129-2518-45f4-9f34-a4fcd4d859af",
		Name:            "Kodama of the North Tree",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "shroud"},
	})
}
