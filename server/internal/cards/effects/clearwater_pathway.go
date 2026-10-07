package effects

// Clearwater Pathway // Murkwater Pathway — modal double-faced Land:
//
//	Clearwater Pathway: "{T}: Add {U}."
//	Murkwater Pathway: "{T}: Add {B}."
//
// This file is the FRONT face (face 0, the bare oracle ID). The back
// face, Murkwater Pathway, was already registered in mdfc_lands.go
// under "<oracle_id>#1" with the rest of the Kaldheim Pathway cycle,
// and nothing registered the front: a land with no basic land type
// derives no mana from its type line, so the blue side tapped for
// nothing. Neither face enters tapped and neither has any other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "144119bc-7fd1-45c5-9e29-f742e7c255ac",
		Name:         "Clearwater Pathway",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			backManaAbility([]string{"U"}),
		},
	})
}
