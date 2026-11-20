package effects

// Knotvine Mystic — Creature — Elf Druid {R}{G}{W}, 2/2 (Conflux,
// #5587):
//
//	"{1}, {T}: Add {R}{G}{W}."
//
// Crystal Quarry's fixed-string mana ability at Naya's three colours.
// The {T} is held back by summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d045695-e4b3-4daa-bfab-d77640fedfd6",
		Name:         "Knotvine Mystic",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{R}{G}{W}",
			Label:    "{1}, {T}: Add {R}{G}{W}",
		}},
	})
}
