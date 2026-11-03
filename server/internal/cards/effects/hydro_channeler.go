package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hydro-Channeler — Creature — Merfolk Wizard {1}{U}, 1/3 (Secrets of
// Strixhaven, #4456):
//
//	"{T}: Add {U}. Spend this mana only to cast an instant or sorcery spell.
//	 {1}, {T}: Add one mana of any color. Spend this mana only to cast an
//	 instant or sorcery spell."
//
// Two mana abilities (CR 605.1a), each restricted to the cast purpose and an
// any-of Instant/Sorcery type restriction (Pelargir Survivor's tags), so the
// mana cannot pay an activation cost or a creature spell. The second pays {1}
// as part of its cost. Summoning sickness gates both taps (CR 302.6).
//
// No simplification.
func init() {
	restrictions := []string{ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")}
	Register(Spec{
		OracleID:     "95630a83-3699-4340-b3e5-17e81db1b82f",
		Name:         "Hydro-Channeler",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{U}",
				Restrictions: restrictions,
				Label:        "Add {U}. Spend this mana only to cast an instant or sorcery spell",
			},
			{
				Cost:         ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced:     "{W|U|B|R|G}",
				Restrictions: restrictions,
				Label:        "{1}, {T}: Add one mana of any color. Spend this mana only to cast an instant or sorcery spell",
			},
		},
	})
}
