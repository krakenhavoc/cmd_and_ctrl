package effects

// Blightsoil Druid — Creature — Elf Druid {1}{B}, 1/2 (Morningtide, #5592):
//
//	"{T}, Pay 1 life: Add {G}."
//
// A mana ability (CR 605.1a) whose cost is a tap and 1 life, the Deserts'
// life-cost row (cmdr2B13DesertLand) on a creature; the engine refuses the
// tap while the Druid is summoning sick (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "351091cb-1263-4202-9447-021c6019cff1",
		Name:         "Blightsoil Druid",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Life: 1},
			Produced: "{G}",
			Label:    "{T}, Pay 1 life: Add {G}",
		}},
	})
}
