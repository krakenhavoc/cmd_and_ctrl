package effects

// Moltensteel Dragon — Artifact Creature — Phyrexian Dragon {4}{R/P}{R/P}, 4/4 (New Phyrexia, #5575):
//
//	"({R/P} can be paid with either {R} or 2 life.)
//	 Flying
//	 {R/P}: This creature gets +1/+0 until end of turn."
//
// Flying and the shared firebreathing row (thisGetsRow). Each Phyrexian
// symbol, in the mana cost and the row's cost, is paid with its colour or
// 2 life (CR 107.4f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dc2d67bb-29b3-43d7-8ebf-2100c5287bbf",
		Name:            "Moltensteel Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{thisGetsRow("{R/P}", 1, 0, "Moltensteel Dragon")},
	})
}
