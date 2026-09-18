package effects

// Manakin — Artifact Creature — Construct {2}, 1/1 (EDHREC rank
// 4224):
//
//	"{T}: Add {C}."
//
// A two-mana mana rock that is also a creature, which is the entire
// reason it is played over a Mind Stone: it is an artifact creature
// for Urza, an Elf-adjacent body for Chief of the Foundry, a sacrifice
// for Krark-Clan Ironworks, and a target for Birthing Pod.
//
// Being a creature is also its drawback and it costs nothing to model:
// summoning sickness gates the tap ability on its own (CR 302.6 — a
// creature's {T} ability needs it to have been controlled since the
// player's most recent turn began), and a board wipe kills it. Both
// fall straight out of the printed type line, so the Spec is one mana
// ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2343af0-468b-42bc-8a0c-347c10f7e2f3",
		Name:         "Manakin",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
