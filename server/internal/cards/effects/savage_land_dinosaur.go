package effects

// Savage Land Dinosaur — {4}{G}{G} Creature — Dinosaur, 7/6 (Marvel Super
// Heroes, #4455):
//
//	"Trample
//	 Basic landcycling {2}"
//
// Apex Witchstalker's landcycling row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6e21ebdf-b97c-494e-a1d5-19db1b7d3016",
		Name:            "Savage Land Dinosaur",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
