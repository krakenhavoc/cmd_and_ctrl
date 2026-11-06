package effects

// Malevolent Chandelier — Artifact Creature — Construct {6}, 4/4
// (Duskmourn: House of Horror, #4466):
//
//	"Flying
//	 {2}: Put target card from a graveyard on the bottom of its owner's
//	 library. Activate only as a sorcery."
//
// The activated ability is Sundering Archaic's graveyard tuck (any
// graveyard's card, to the bottom of its OWNER's library), held to
// sorcery timing (CR 602.5d): your main phase, with an empty stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "437553c1-0aa4-4093-b6e5-c175ede5f278",
		Name:            "Malevolent Chandelier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:        "{2}: Put target card from a graveyard on the bottom of its owner's library. Activate only as a sorcery.",
			Cost:         ManaCost("{2}"),
			Targets:      TargetCardInGraveyard("target card from a graveyard"),
			SorcerySpeed: true,
			Effect:       socB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
