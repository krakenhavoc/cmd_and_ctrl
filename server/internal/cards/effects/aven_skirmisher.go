package effects

// Aven Skirmisher — Creature — Bird Warrior {W}, 1/1 (Fate Reforged,
// #5552):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "57d8d656-ccbd-4deb-abf1-a7fde4e7d691",
		Name:            "Aven Skirmisher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
