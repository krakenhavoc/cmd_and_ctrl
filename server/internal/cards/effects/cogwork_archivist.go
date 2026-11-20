package effects

// Cogwork Archivist — Artifact Creature — Construct {6}, 4/5 (Strixhaven:
// School of Mages, #5502):
//
//	"Reach
//	 {2}, {T}: Put target card from a graveyard on the bottom of its
//	 owner's library."
//
// Reach is the printed keyword; the row is Chrome Companion's, any
// graveyard's card to the bottom of its owner's library.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60674ef7-f8e2-41e8-bab3-0100cfc9fbe9",
		Name:            "Cogwork Archivist",
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
