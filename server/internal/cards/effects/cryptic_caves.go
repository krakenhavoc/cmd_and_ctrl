package effects

// Cryptic Caves — Land (EDHREC rank 4374):
//
//	"{T}: Add {C}.
//	 {1}, {T}, Sacrifice this land: Draw a card. Activate only if you
//	 control five or more lands."
//
// Mind Stone's draw on a land, for the late game. The gate is the
// activation condition (CR 602.1b, #743), ControlsAtLeast(5, lands),
// and the Caves count themselves, as Temple of the False God does.
// Checked as the ability is activated; the sacrifice then leaves four,
// which does not matter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2eb7a64-a307-4a78-a25d-63fb3ae1e237",
		Name:         "Cryptic Caves",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:     "{1}, {T}, Sacrifice this land: Draw a card. Activate only if you control five or more lands.",
			Cost:      Plus(ManaCost("{1}"), TapCost(), SacrificeThis()),
			Condition: ControlsAtLeast(5, MatchLand),
			Effect:    b36DrawOne,
		}},
	})
}
