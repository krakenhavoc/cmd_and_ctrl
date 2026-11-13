package effects

// Skyshroud Troopers — Creature — Elf Druid Warrior {3}{G}, 3/3
// (Stronghold, #5627):
//
//	"{T}: Add {G}."
//
// Llanowar Elves' mana ability on a bigger body: one fixed {G}, added at
// once without a colour prompt, and off the stack (CR 605.3b). The tap
// symbol means it can't be activated the turn the Troopers arrive unless
// they have haste (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5a40cd8-e1f9-4294-99ef-26fb455d4f95",
		Name:         "Skyshroud Troopers",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
