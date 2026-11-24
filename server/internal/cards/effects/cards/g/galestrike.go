package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galestrike — Instant {2}{U} (Commander Legends):
//
//	"Return target tapped creature to its owner's hand.
//	 Draw a card."
//
// Repulse's body over a tapped-creature target: the target is re-checked
// on resolution (CR 608.2b), so a creature that untapped or left in
// response makes the whole spell do nothing, the draw included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3006d992-2714-4264-a4c3-953b67dc7347",
		Name:         "Galestrike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
		OnResolve:    effects.BounceTargetThenDrawACard,
	})
}
