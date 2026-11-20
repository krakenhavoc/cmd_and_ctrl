package effects

// Oboro Breezecaller — Creature — Moonfolk Wizard {1}{U}, 1/1 (Saviors of
// Kamigawa, #5601):
//
//	"Flying
//	 {2}, Return a land you control to its owner's hand: Untap target
//	 land."
//
// The return is Meloku the Clouded Mirror's cost, the land named as the
// ability is activated; the untap is Deserted Temple's effect. A land
// returned for the cost is gone before the target is checked, so it
// cannot be the land untapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "257ee9a0-5ee0-49c4-8619-e75e1134d7ac",
		Name:            "Oboro Breezecaller",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label: "{2}, Return a land you control to its owner's hand: Untap target land.",
			Cost: Plus(
				ManaCost("{2}"),
				ReturnAPermanentToHand("a land you control", Land()),
			),
			Targets: TargetPermanent("target land", Land()),
			Effect:  UntapEachLegalTarget,
		}},
	})
}
