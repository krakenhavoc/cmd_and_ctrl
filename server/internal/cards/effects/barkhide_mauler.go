package effects

// Barkhide Mauler — Creature — Beast {4}{G}, 4/4 (Onslaught):
//
//	"Cycling {2}"
//
// A vanilla 4/4 whose one ability is the keyword's hand ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "84a18b1f-5f9f-4e2b-87cd-c437bdf4d278",
		Name:         "Barkhide Mauler",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
