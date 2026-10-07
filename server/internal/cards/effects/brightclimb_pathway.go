package effects

// Brightclimb Pathway // Grimclimb Pathway — modal double-faced Land:
//
//	Brightclimb Pathway: "{T}: Add {W}."
//	Grimclimb Pathway: "{T}: Add {B}."
//
// This file is the FRONT face (face 0, the bare oracle ID). The back
// face, Grimclimb Pathway, was already registered in mdfc_lands.go
// under "<oracle_id>#1" with the rest of the Kaldheim Pathway cycle,
// and nothing registered the front: a land with no basic land type
// derives no mana from its type line, so the white side tapped for
// nothing. Neither face enters tapped and neither has any other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c633e02-95ef-445e-b4e0-fbfbc5ed9cc9",
		Name:         "Brightclimb Pathway",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			backManaAbility([]string{"W"}),
		},
	})
}
