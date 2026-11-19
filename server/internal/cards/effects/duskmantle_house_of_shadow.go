package effects

// Duskmantle, House of Shadow — Land (The List):
//
//	"{T}: Add {C}.
//	 {U}{B}, {T}: Target player mills a card."
//
// The colourless ability is a mana ability; the mill targets (any player,
// the controller included) and is an ordinary activated ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "67b2cd0c-ecc8-4129-b1ac-820c9924190c",
		Name:         "Duskmantle, House of Shadow",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{U}{B}, {T}: Target player mills a card.",
			Cost:    Plus(ManaCost("{U}{B}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect:  s82BNSB08MillTargetPlayer,
		}},
	})
}
