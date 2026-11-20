package effects

// Lifespring Druid — Creature — Elf Druid {2}{G}, 2/1 (Battle for
// Zendikar, #5547):
//
//	"{T}: Add one mana of any color."
//
// Xenobotanist's mana ability: one mana of the colour its controller
// names as it is activated (CR 605.1a), and the {T} cost waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "165c1b17-2bde-4c96-892d-4a8fdc7a57b7",
		Name:         "Lifespring Druid",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
