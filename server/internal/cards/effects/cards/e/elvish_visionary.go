package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Visionary — Creature — Elf Shaman {1}{G}, 1/1 (EDHREC rank
// 2296):
//
//	"When this creature enters, draw a card."
//
// The Elf cantrip. Wall of Omens's ETB on an Elf body: the trigger
// goes on the stack and the controller draws when it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6a3a882-a127-4590-93d7-679ef4313efe",
		Name:         "Elvish Visionary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Elvish Visionary — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
