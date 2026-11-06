package effects

// Gangly Stompling — Creature — Shapeshifter {2}{R/G}, 4/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Changeling (This card is every creature type.)
//	 Trample"
//
// Two keywords the engine honours, declared the way Ambush Viper declares
// its pair: changeling makes it every creature type wherever it is (CR
// 702.73a), and trample is the combat keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60206d4e-4301-4d7a-a68e-46a2ad9fc407",
		Name:            "Gangly Stompling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"changeling", "trample"},
	})
}
