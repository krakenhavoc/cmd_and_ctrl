package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ixalli's Lorekeeper — Creature — Human Druid {G}, 1/1 (The Lost Caverns
// of Ixalan, #4470):
//
//	"{T}: Add one mana of any color. Spend this mana only to cast a
//	 Dinosaur spell or activate an ability of a Dinosaur source."
//
// Jasmine Dragon Tea Shop's restricted any-colour row with Dinosaur for
// Ally: the spend restriction is a disjunction of "cast a Dinosaur spell"
// and "activate an ability of a Dinosaur source", read off the spell or
// the source as the mana is spent (CR 106.6). It is a creature, so the
// {T} ability needs it to have been under its controller's control since
// the turn began (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3ce0fa9-8742-4e2c-a8a0-00fc009f998a",
		Name:         "Ixalli's Lorekeeper",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color. Spend this mana only to cast a Dinosaur spell or activate an ability of a Dinosaur source.",
			Restrictions: []string{game.ManaRestrictAnyOf(
				[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Dinosaur")},
				[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Dinosaur")},
			)},
		}},
	})
}
