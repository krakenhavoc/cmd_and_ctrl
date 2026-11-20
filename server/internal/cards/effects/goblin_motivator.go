package effects

// Goblin Motivator — Creature — Goblin Warrior {R}, 1/1
// (Core Set 2019, #5527):
//
//	"{T}: Target creature gains haste until end of turn. (It can attack and {T} this turn.)"
//
// Bloodlust Inciter's text: a tap ability granting haste to a target
// until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df67597f-42a8-4fa7-b426-d7db8b384445",
		Name:         "Goblin Motivator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature gains haste until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    TapCost(),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("haste", "Goblin Motivator — haste until end of turn"),
		}},
	})
}
