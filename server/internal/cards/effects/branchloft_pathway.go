package effects

// Branchloft Pathway // Boulderloft Pathway — modal double-faced Land:
//
//	Branchloft Pathway: "{T}: Add {G}."
//	Boulderloft Pathway: "{T}: Add {W}."
//
// This file is the FRONT face (face 0, the bare oracle ID). The back
// face, Boulderloft Pathway, was already registered in mdfc_lands.go
// under "<oracle_id>#1" with the rest of the Kaldheim Pathway cycle,
// and nothing registered the front: a land with no basic land type
// derives no mana from its type line, so the green side tapped for
// nothing. Neither face enters tapped and neither has any other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c304547-a4b1-46c9-baed-16d2bfbe16eb",
		Name:         "Branchloft Pathway",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			backManaAbility([]string{"G"}),
		},
	})
}
