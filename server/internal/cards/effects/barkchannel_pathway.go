package effects

// Barkchannel Pathway // Tidechannel Pathway — modal double-faced Land:
//
//	Barkchannel Pathway: "{T}: Add {G}."
//	Tidechannel Pathway: "{T}: Add {U}."
//
// This file is the FRONT face (face 0, the bare oracle ID). The back
// face, Tidechannel Pathway, was already registered in mdfc_lands.go
// under "<oracle_id>#1" with the rest of the Kaldheim Pathway cycle,
// and nothing registered the front: a land with no basic land type
// derives no mana from its type line, so the green side tapped for
// nothing. Neither face enters tapped and neither has any other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59d22de5-e310-44d7-89cf-ef3529e40cef",
		Name:         "Barkchannel Pathway",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			backManaAbility([]string{"G"}),
		},
	})
}
