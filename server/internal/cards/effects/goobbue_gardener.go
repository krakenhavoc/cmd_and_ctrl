package effects

// Goobbue Gardener — Creature — Plant Beast {1}{G}, 1/3 (Final Fantasy,
// #4462):
//
//	"{T}: Add {G}."
//
// Llanowar Elves's mana ability on a bigger body: a single fixed {G}, so
// it adds the mana at once. The tap cost is a creature's, so it waits out
// summoning sickness (CR 302.6) like any mana dork.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5c64331-2ba5-43ad-b77c-979875cc39e3",
		Name:         "Goobbue Gardener",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
