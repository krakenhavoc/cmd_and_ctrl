package effects

// Prahv, Spires of Order — Land:
//
//	"{T}: Add {C}.
//	 {4}{W}{U}, {T}: Prevent all damage a source of your choice would deal this turn."
//
// ADR 0108 §7 (#1904): a shield against a source chosen as the ability
// resolves (CR 609.7a), preventing all of that source's damage, to
// anything, for the rest of the turn.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "37ff5ba6-0763-4c73-85bf-66856e67b8f3",
		Name:         "Prahv, Spires of Order",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{sourceShieldRow(
			"{4}{W}{U}, {T}: Prevent all damage a source of your choice would deal this turn.",
			Plus(ManaCost("{4}{W}{U}"), TapCost()), nil,
			PreventDamageFromChosenSource(ShieldAnything))},
	})
}
