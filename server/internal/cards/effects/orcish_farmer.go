package effects

// Orcish Farmer — Creature — Orc {1}{R}{R}, 2/2:
//
//	"{T}: Target land becomes a Swamp until its controller's next untap
//	 step."
//
// ADR 0109 §1 decision 5 (#1881): CR 305.7's type set, lasting until the
// land's controller's next turn begins. The untap step is the first step
// of a turn (CR 501.1), and the engine ends an "until your next
// turn" effect as that turn begins, before anything untaps, so the land is
// a Swamp through every other player's turn and is itself again when its
// controller untaps. "Its controller" is read as the ability resolves.
// Until then the land's land types are replaced by Swamp (other subtypes
// stay, CR 205.1a), it loses its rules-text abilities and taps for {B}
// (CR 305.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3039d19-8c98-4953-8943-9922b6ab45ef",
		Name:         "Orcish Farmer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target land becomes a Swamp until its controller's next untap step.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target land", Land()),
			Effect:  TargetLandBecomesUntilItsControllersNextTurn("Orcish Farmer", "Swamp"),
		}},
	})
}
