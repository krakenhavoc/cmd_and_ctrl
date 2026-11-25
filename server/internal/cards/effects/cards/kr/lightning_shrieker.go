package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightning Shrieker — Creature — Dragon {4}{R}, 5/5 (Jumpstart, #5510):
//
//	"Flying, trample, haste
//	 At the beginning of the end step, this creature's owner shuffles it into
//	 their library."
//
// Every end step, not only its controller's. The owner shuffles it in, so a
// Shrieker you stole goes back to its owner's library; it leaves if it is
// still the same permanent as the trigger resolves (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a5196310-f449-4d5e-8ea8-901a83bad6ff",
		Name:            "Lightning Shrieker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AnyPlayer, "Lightning Shrieker — its owner shuffles it into their library",
				effects.R35CeruleanSphinxShuffleItself),
		},
	})
}
