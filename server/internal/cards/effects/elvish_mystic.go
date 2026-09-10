package effects

// Elvish Mystic — Creature — Elf Druid {G}, 1/1 (EDHREC rank 103):
//
//	"{T}: Add {G}."
//
// Llanowar Elves with a different name and oracle ID. Green decks
// that want the effect want four of it and settle for two, which is
// why both are top-of-list in a singleton format.
//
// Summoning sickness applies, as on Llanowar Elves (CR 302.1,
// enforced by ActivateManaAbility since #233).
func init() {
	Register(Spec{
		OracleID: "3f3b2c10-21f8-4e13-be83-4ef3fa36e123",
		Name:     "Elvish Mystic",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
