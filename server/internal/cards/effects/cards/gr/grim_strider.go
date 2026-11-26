package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Strider — {3}{B} Creature — Horror 6/6 (Amonkhet, S82 old batch 246):
//
//	"This creature gets -1/-1 for each card in your hand."
//
// A layer 7c modification (CR 613.4c) of the source alone, recomputed on every
// change to its controller's hand (DependsOnHandSize) as Stingerback Terror's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9dcdf5b-0fdb-4d8e-8ddf-08e44f67018d",
		Name:         "Grim Strider",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.SelfMinusOnePerCardInYourHand()},
	})
}
