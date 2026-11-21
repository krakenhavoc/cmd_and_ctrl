package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marauding Boneslasher — {2}{B} Creature — Zombie Minotaur 3/3 (Jumpstart 2022, #5948):
//
//	"This creature can't block unless you control another Zombie."
//
// A conditional can't-block rule (CR 509.1b) on the carrier alone: it is
// refused as a blocker while its controller controls no other Zombie.
// "Another" is object identity and the Zombies are read after the layers
// (hobA9OtherSubtypeYouControl), so a changeling counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0bc9c88-8bec-4d92-a078-08883c82cb28",
		Name:         "Marauding Boneslasher",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{{
			Reason: game.BlockReasonCantBlockAttacker,
			Label:  "it can't block unless its controller controls another Zombie",
			Pair: func(g *game.Game, _, blocker, source *game.Card) bool {
				return blocker.InstanceID == source.InstanceID && effects.HobA9OtherSubtypeYouControl(g, source, "Zombie") == 0
			},
		}},
	})
}
