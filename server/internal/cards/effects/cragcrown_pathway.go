package effects

// Cragcrown Pathway // Timbercrown Pathway — modal double-faced Land:
//
//	Cragcrown Pathway: "{T}: Add {R}."
//	Timbercrown Pathway: "{T}: Add {G}."
//
// This file is the FRONT face (face 0, the bare oracle ID). The back
// face, Timbercrown Pathway, was already registered in mdfc_lands.go
// under "<oracle_id>#1" with the rest of the Kaldheim Pathway cycle,
// and nothing registered the front: a land with no basic land type
// derives no mana from its type line, so the red side tapped for
// nothing. Neither face enters tapped and neither has any other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "727ca426-f4cc-4218-8ae5-8c427af2e816",
		Name:         "Cragcrown Pathway",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			backManaAbility([]string{"R"}),
		},
	})
}
