package effects

// Needleverge Pathway // Pillarverge Pathway — modal double-faced Land:
//
//	Needleverge Pathway: "{T}: Add {R}."
//	Pillarverge Pathway: "{T}: Add {W}."
//
// This file is the FRONT face (face 0, the bare oracle ID). The back
// face, Pillarverge Pathway, was already registered in mdfc_lands.go
// under "<oracle_id>#1" with the rest of the Kaldheim Pathway cycle,
// and nothing registered the front: a land with no basic land type
// derives no mana from its type line, so the red side tapped for
// nothing. Neither face enters tapped and neither has any other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a9b8d020-4d72-4934-8942-df29ef19fc1d",
		Name:         "Needleverge Pathway",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			backManaAbility([]string{"R"}),
		},
	})
}
