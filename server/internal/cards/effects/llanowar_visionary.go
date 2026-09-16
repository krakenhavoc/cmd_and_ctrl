package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Llanowar Visionary — Creature — Elf Druid {2}{G}, 2/2 (EDHREC rank
// 3755):
//
//	"When this creature enters, draw a card.
//	 {T}: Add {G}."
//
// Elvish Visionary stapled to a Llanowar Elves. The draw is a real
// ETB trigger with a response window; the mana ability is a {T} on
// a creature, so summoning sickness applies (CR 302.6) — the engine
// enforces it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f75ed312-3a23-4624-80c5-03980aa22d0b",
		Name:         "Llanowar Visionary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Llanowar Visionary — draw a card", b36DrawOne),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
