package effects

// Kree Sentinel — Artifact Creature — Kree Robot Villain {4}{R}, 5/5
// (Marvel Super Heroes, #4455):
//
//	"Reach
//	 Basic landcycling {2}"
//
// Apex Witchstalker's shape: reach is the printed keyword and basic
// landcycling is the shared typecycling constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c2b9bb2-8201-47e8-84d5-f3518cac03cb",
		Name:            "Kree Sentinel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
