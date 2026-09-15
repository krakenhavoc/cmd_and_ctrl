package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c6a3a882-a127-4590-93d7-679ef4313efe",
		Name:         "Elvish Visionary",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elvish Visionary — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
