package effects

// mana_myr.go — the Mirrodin "mana Myr" cycle:
//
//	"{T}: Add {R}."
//
// Artifact Creature — Myr {2}, 1/1, one per colour. A Llanowar Elf
// that is also an artifact, which is why the artifact decks run
// them. Roadmap batch 11 (#304) ranks Iron Myr and Silver Myr, batch
// 12 (#305) Gold Myr; Leaden and Copper Myr belong in this table
// when their batches reach them. Palladium Myr ({T}: Add {C}{C}) is
// its own file.
//
// The tap is on a CREATURE, so summoning sickness applies (CR
// 302.1); ActivateManaAbility enforces that from the type line.
//
// No simplification.
func init() {
	for _, myr := range []struct{ oracleID, name, color string }{
		{"6c5cbab6-ee27-46f5-97a7-df85698d1e9f", "Iron Myr", "R"},
		{"66e8f7f8-3a6d-46ba-837c-b9713ddf7f40", "Silver Myr", "U"},
		// Roadmap batch 12 (#305).
		{"bd6af7b3-b30f-4a65-a18f-8655f778e76a", "Gold Myr", "W"},
	} {
		Register(Spec{
			OracleID:     myr.oracleID,
			Name:         myr.name,
			Completeness: CompletenessFull,
			ManaAbilities: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{" + myr.color + "}",
				Label:    "Add {" + myr.color + "}",
			}},
		})
	}
}
