package effects

// Ledev Guardian — Creature — Human Knight {3}{W}, 2/4 (Guilds of Ravnica, #5525):
//
//	"Convoke"
//
// Convoke is the shared tap cost (CR 702.51a), as on Halo Hopper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "edcf36b4-dea4-4314-94da-e0ee5e30cd37",
		Name:         "Ledev Guardian",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
	})
}
