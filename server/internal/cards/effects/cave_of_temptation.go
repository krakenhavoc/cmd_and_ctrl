package effects

// Cave of Temptation — Land (Foundations Jumpstart, #5929):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {4}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 creature. Activate only as a sorcery."
//
// The same three abilities as Captivating Cave.
func init() {
	Register(Spec{
		OracleID:     "75540897-53f6-433b-bd70-9851551df6ef",
		Name:         "Cave of Temptation",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			painlessColorless(),
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Activated: []ActivatedAbility{{
			Label:        "{4}, {T}, Sacrifice this land: Put two +1/+1 counters on target creature. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature"),
			Effect:       b31CountersOnChosenCreature,
		}},
	})
}
