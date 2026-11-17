package effects

// Phyrexian Archivist — Artifact Creature — Phyrexian Construct 4/5 {6} (March of the Machine, S82):
//
//	"Reach
//	 {2}, {T}: Put target card from a graveyard on the bottom of its owner's library."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cdef800d-033e-4095-964d-9b6066f9fd36",
		Name:            "Phyrexian Archivist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  socB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
