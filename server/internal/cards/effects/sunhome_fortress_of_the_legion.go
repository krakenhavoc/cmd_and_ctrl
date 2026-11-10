package effects

// Sunhome, Fortress of the Legion — Land:
//
//	"{T}: Add {C}.
//	 {2}{R}{W}, {T}: Target creature gains double strike until end of
//	 turn."
//
// Axgard Cavalry's grant on a land: the keyword is granted to the
// target as the ability resolves and lasts until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a9f8344c-1705-4254-81d6-aa05e0c69c29",
		Name:          "Sunhome, Fortress of the Legion",
		Completeness:  CompletenessFull,
		ManaAbilities: []ManaAbility{painlessColorless()},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}{W}, {T}: Target creature gains double strike until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "double strike")),
			Cost:    Plus(ManaCost("{2}{R}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("double strike", "Sunhome, Fortress of the Legion — double strike until end of turn"),
		}},
	})
}
