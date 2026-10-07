package effects

// Blightstep Pathway // Searstep Pathway — modal double-faced Land:
//
//	Blightstep Pathway: "{T}: Add {B}."
//	Searstep Pathway: "{T}: Add {R}."
//
// This file is the FRONT face (face 0, the bare oracle ID). The back
// face, Searstep Pathway, was already registered in mdfc_lands.go
// under "<oracle_id>#1" with the rest of the Kaldheim Pathway cycle,
// and nothing registered the front: a land with no basic land type
// derives no mana from its type line, so the black side tapped for
// nothing. Neither face enters tapped and neither has any other text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e580a229-e800-4746-9d37-c32fcef8de28",
		Name:         "Blightstep Pathway",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			backManaAbility([]string{"B"}),
		},
	})
}
