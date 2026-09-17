package effects

// Bonders' Enclave — Land (EDHREC rank 544):
//
//	"{T}: Add {C}.
//	 {3}, {T}: Draw a card. Activate only if you control a creature
//	 with power 4 or greater."
//
// A colorless land that draws in a big-creature deck. The gate is the
// activation condition (CR 602.1b, #743), ControlsAtLeast(1, …) over
// the creature's CURRENT power — an anthem or +1/+1 counters count, as
// the printed "power 4 or greater" does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f33ce38a-34ec-4b65-a0fc-160484a02007",
		Name:         "Bonders' Enclave",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{3}, {T}: Draw a card. Activate only if you control a creature with power 4 or greater.",
			Cost:      Plus(ManaCost("{3}"), TapCost()),
			Condition: ControlsAtLeast(1, MatchCreatureWithPowerAtLeast(4)),
			Effect:    b36DrawOne,
		}},
	})
}
