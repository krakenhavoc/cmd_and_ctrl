package effects

// Llanowar Tribe — Creature — Elf Druid {G}{G}{G}, 3/3 (EDHREC rank
// 2244):
//
//	"{T}: Add {G}{G}{G}."
//
// Three Llanowar Elves in one body. A creature mana ability with a
// tap cost, so summoning sickness applies (CR 302.1) — the engine
// enforces it inside ActivateManaAbility. The mana is three green
// slots, no pick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfd0f0e6-1bbf-4450-97d2-c54907abb7e4",
		Name:         "Llanowar Tribe",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}{G}{G}",
			Label:    "Add {G}{G}{G}",
		}},
	})
}
