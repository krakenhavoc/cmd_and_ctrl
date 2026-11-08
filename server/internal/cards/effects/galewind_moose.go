package effects

// Galewind Moose — Creature — Elemental Elk {4}{G}{G}, 6/6 (Bloomburrow,
// #4467):
//
//	"Flash
//	 Reach, vigilance, trample"
//
// Four printed keywords and nothing else, declared as Ambush Viper
// declares its two: flash lets it be cast any time you could cast an
// instant (CR 702.8a), and reach, vigilance and trample are the engine's
// combat keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b0747a4f-d041-4750-aef5-a5d101696f65",
		Name:            "Galewind Moose",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "reach", "vigilance", "trample"},
	})
}
