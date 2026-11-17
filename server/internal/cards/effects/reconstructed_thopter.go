package effects

// Reconstructed Thopter — Artifact Creature — Thopter {3}, 2/1:
//
//	"Flying
//	 Unearth {2}"
func init() {
	Register(Spec{
		OracleID:        "c87ff8e8-a3da-4848-9881-fbc896acf7c5",
		Name:            "Reconstructed Thopter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{Unearth("{2}")},
	})
}
