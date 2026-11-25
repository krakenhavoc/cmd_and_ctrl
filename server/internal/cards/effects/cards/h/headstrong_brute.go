package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Headstrong Brute — Creature — Orc Pirate {2}{R}, 3/3 (Ixalan, #5532):
//
//	"This creature can't block.
//	 This creature has menace as long as you control another Pirate."
//
// Rimrock Knight's block restriction, and Bolg's Company's self grant
// (layer 6, CR 613.1f): the Brute itself, a Pirate, never satisfies
// "another" (CR 109.1); a Pirate token or any other Pirate permanent you
// control does, and it is read as the layers stand now.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "675ed28e-4328-4638-aaad-3f47e5c7895f",
		Name:         "Headstrong Brute",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.RestrictSelf(game.CantBlock),
			effects.KeywordGrant(effects.HobA9WhileYouControlAnother("Pirate"), "menace"),
		},
	})
}
