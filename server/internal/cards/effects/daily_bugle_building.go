package effects

// Daily Bugle Building — Land (Marvel's Spider-Man):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 Smear Campaign — {1}, {T}: Target legendary creature gains menace until
//	 end of turn. Activate only as a sorcery."
//
// The first two are mana abilities (the second a filter that costs {1}); the
// third is Hanweir Battlements' targeted until-end-of-turn grant at sorcery
// speed. "Smear Campaign" is an ability word, with no rules meaning.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "483e0c6c-8131-486c-b482-cc3396c9786b",
		Name:         "Daily Bugle Building",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
			{Cost: ManaAbilityCost{Mana: "{1}", Tap: true}, Produced: "{W|U|B|R|G}", Label: "{1}, {T}: Add one mana of any color."},
		},
		Activated: []ActivatedAbility{{
			Label:        "Smear Campaign — {1}, {T}: Target legendary creature gains menace until end of turn. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 0, 0, "menace")),
			Cost:         Plus(ManaCost("{1}"), TapCost()),
			SorcerySpeed: true,
			Targets:      TargetCreature("target legendary creature", Legendary()),
			Effect:       ebTargetCreatureGainsUntilEOT("menace", "Daily Bugle Building — menace until end of turn"),
		}},
	})
}
