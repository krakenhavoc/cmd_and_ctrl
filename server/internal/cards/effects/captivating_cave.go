package effects

// Captivating Cave — Land — Cave (The Lost Caverns of Ixalan, #4470):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {4}, {T}, Sacrifice this land: Put two +1/+1 counters on target
//	 creature. Activate only as a sorcery."
//
// The mana abilities are Crystal Grotto's (a colourless tap and the
// {1}, {T} filter into the five-colour pipe). The sacrifice row is The
// Hobbit's counter lands' (hobB7CounterLand) on any creature: two +1/+1
// counters on the creature chosen at announce, if it is still legal as
// the ability resolves (CR 608.2b). The land is gone either way: the
// sacrifice is a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c77767a-8133-43bc-b7a5-09a73259d354",
		Name:         "Captivating Cave",
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
