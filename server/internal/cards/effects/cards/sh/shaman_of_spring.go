package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shaman of Spring — Creature — Elf Shaman {3}{G}, 2/2 (Magic 2015,
// #5556):
//
//	"When this creature enters, draw a card."
//
// Elvish Visionary's enters trigger: it goes on the stack and the
// controller draws when it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b1ee0250-83ea-45a9-b5e5-3fa74bc7f26b",
		Name:         "Shaman of Spring",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shaman of Spring — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
