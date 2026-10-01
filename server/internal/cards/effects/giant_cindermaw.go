package effects

// Giant Cindermaw — Creature — Dinosaur Beast {2}{R}, 4/3:
//
//	"Trample
//	 Players can't gain life."
//
// Trample rides PrintedKeywords; "Players can't gain life" is ADR 0107
// §5's battlefield static (CR 119.7, #1880), its controller included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bd0c0e85-ff24-46c9-a547-4f633355c131",
		Name:            "Giant Cindermaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		CantGainLife:    PlayersCantGainLife(),
	})
}
