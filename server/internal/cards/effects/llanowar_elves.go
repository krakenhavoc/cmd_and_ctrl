package effects

// Llanowar Elves — Creature — Elf Druid {G}, 1/1 (EDHREC rank 58):
//
//	"{T}: Add {G}."
//
// The archetypal one-drop mana dork. Same shape as Birds of
// Paradise, already in the catalog, minus the colour fixing: a
// single fixed {G} needs no pipe, so activating it adds mana
// immediately instead of queueing a colour choice.
//
// Summoning sickness applies: Game.ActivateManaAbility checks CR
// 302.1 (since #233), so the Elf cannot tap for mana the turn it
// lands unless it has haste. This card shipped in #243 documenting
// the opposite as a known gap — the two PRs crossed on main —
// and TestLlanowarElvesCannotTapWhileSummoningSick pins the rule.
func init() {
	Register(Spec{
		OracleID: "68954295-54e3-4303-a6bc-fc4547a4e3a3",
		Name:     "Llanowar Elves",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
