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
// KNOWN ENGINE GAP — summoning sickness is NOT enforced on mana
// abilities. Game.ActivateManaAbility checks only whether the
// permanent is already tapped; it never consults CR 302.1, so this
// Elf taps for mana the turn it lands. The activated-ability path
// added in S21 does check (HasSummoningSickness → ErrSummoningSick),
// so the two ability kinds currently disagree about the same rule.
//
// The gap predates this card — Birds of Paradise has had it since
// S15 — and the fix is three lines inside ActivateManaAbility, left
// out of this batch only because an open PR is editing that
// function. TestLlanowarElvesTapsWhileSummoningSick_KnownGap pins
// the current behaviour and is the fix's acceptance test.
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
