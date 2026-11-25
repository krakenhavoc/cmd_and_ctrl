package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tome of the Guildpact — Artifact — Book {5} (Ravnica Allegiance):
//
//	"Whenever you cast a multicolored spell, draw a card.
//	 {T}: Add one mana of any color."
//
// A spell is multicolored when it has two or more colors, read on the stack
// (CR 105.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c7dd86d-25eb-4b24-9358-98b951248e04",
		Name:         "Tome of the Guildpact",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Multicolored(), "Tome of the Guildpact — draw a card",
				effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
