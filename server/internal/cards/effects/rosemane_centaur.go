package effects

// Rosemane Centaur — Creature — Centaur Soldier {3}{G}{W}, 4/4 (Guilds of Ravnica, #5525):
//
//	"Convoke
//	 Vigilance"
//
// Sun-Dappled Celebrant's shape: convoke is the shared tap cost (CR
// 702.51a) and vigilance is printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "54b3ccdc-e5b0-44d9-bc81-a296fc5dc005",
		Name:            "Rosemane Centaur",
		Completeness:    CompletenessFull,
		TapCost:         Convoke(),
		PrintedKeywords: []string{"vigilance"},
	})
}
