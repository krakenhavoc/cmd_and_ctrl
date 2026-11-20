package effects

// Highborn Ghoul — Creature — Zombie {B}{B}, 2/1 (Dark Ascension, #5571):
//
//	"Intimidate (This creature can't be blocked except by artifact
//	 creatures and/or creatures that share a color with it.)"
//
// Intimidate is an engine keyword carried as printed data (CR 702.13), as
// on Bladetusk Boar.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c145f2d-8345-4902-b869-afb70145dec7",
		Name:            "Highborn Ghoul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
	})
}
