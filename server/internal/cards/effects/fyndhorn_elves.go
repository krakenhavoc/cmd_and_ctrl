package effects

// Fyndhorn Elves — Creature — Elf Druid {G}, 1/1 (EDHREC rank 215):
//
//	"{T}: Add {G}."
//
// Llanowar Elves with a different name — the second copy every
// green deck runs. Same single fixed {G}, so activating it adds the
// mana immediately with no colour prompt, and the same CR 302.6
// summoning-sickness gate the engine applies to every tap ability on
// a creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "df317532-7d36-40fd-938f-e972749c8792",
		Name:         "Fyndhorn Elves",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
