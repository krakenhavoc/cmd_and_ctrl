package effects

// Birds of Paradise — 0/1 Creature — Bird with "Flying" and
// "{T}: Add one mana of any color."
//
// Both halves are implemented. The mana ability went live with the
// S15 mana pipeline; flying became enforceable with the S18 keyword
// pipeline and is declared below.
//
// The flying declaration was missing until #338's stale-simplification
// sweep: the file still carried an S14 note calling the keyword
// "cosmetic", and nobody went back for it once S18 landed. Since
// #330 a deck-imported copy picks flying up from Scryfall via
// Card.Keywords, which hid the gap — but a Birds that never goes
// through deck import (fixture, demo seed, token copy) has only the
// catalog to read, and blocked nothing.
func init() {
	Register(Spec{
		OracleID:        "d3a0b660-358c-41bd-9cd2-41fbf3491b1a",
		Name:            "Birds of Paradise",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
