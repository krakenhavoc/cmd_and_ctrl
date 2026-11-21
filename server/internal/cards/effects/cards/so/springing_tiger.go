package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Springing Tiger — Creature — Cat {3}{G}, 3/3 (Odyssey, #5613):
//
//	"Threshold — This creature gets +2/+2 as long as there are seven or
//	 more cards in your graveyard."
//
// Krosan Beast's threshold at +2/+2: the layer 7c self-modifier (CR
// 613.4c), read live on every recompute.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f8c4805-694d-4408-8eb5-9fb7ce278ded",
		Name:         "Springing Tiger",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.FraThresholdSelfPT(2, 2)},
	})
}
