package effects

// Riverglide Pathway // Lavaglide Pathway — modal double-faced Land:
//
//	Riverglide Pathway: "{T}: Add {U}."
//	Lavaglide Pathway: "{T}: Add {R}."
//
// This file is the FRONT face (face 0, the bare oracle ID). The back
// face, Lavaglide Pathway, was already registered in mdfc_lands.go
// under "<oracle_id>#1" with the rest of the Kaldheim Pathway cycle,
// and nothing registered the front: a land with no basic land type
// derives no mana from its type line, so the blue side tapped for
// nothing. Neither face enters tapped and neither has any other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4924b3a4-a218-4783-8a4d-82361fdecc78",
		Name:         "Riverglide Pathway",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			backManaAbility([]string{"U"}),
		},
	})
}
