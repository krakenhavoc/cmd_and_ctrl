package effects

// Hengegate Pathway // Mistgate Pathway — modal double-faced Land:
//
//	Hengegate Pathway: "{T}: Add {W}."
//	Mistgate Pathway: "{T}: Add {U}."
//
// This file is the FRONT face (face 0, the bare oracle ID). The back
// face, Mistgate Pathway, was already registered in mdfc_lands.go
// under "<oracle_id>#1" with the rest of the Kaldheim Pathway cycle,
// and nothing registered the front: a land with no basic land type
// derives no mana from its type line, so the white side tapped for
// nothing. Neither face enters tapped and neither has any other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "461b3f2f-fcee-4160-abfa-061f8b6a784f",
		Name:         "Hengegate Pathway",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			backManaAbility([]string{"W"}),
		},
	})
}
