package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krosan Beast — Creature — Squirrel Beast {3}{G}, 1/1 (Odyssey, #5613):
//
//	"Threshold — This creature gets +7/+7 as long as there are seven or
//	 more cards in your graveyard."
//
// Threshold is an ability word with no rules of its own (CR 207.2c): the
// layer 7c self-modifier (CR 613.4c) Mind Drill Assailant uses, read live
// on every recompute, so the Beast is an 8/8 while its controller's
// graveyard holds seven or more cards and a 1/1 again the moment it
// doesn't.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "042a665d-2f63-4f7e-ac89-0d560da68085",
		Name:         "Krosan Beast",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.FraThresholdSelfPT(7, 7)},
	})
}
