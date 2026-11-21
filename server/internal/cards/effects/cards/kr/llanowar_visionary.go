package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f75ed312-3a23-4624-80c5-03980aa22d0b",
		Name:         "Llanowar Visionary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Llanowar Visionary — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
