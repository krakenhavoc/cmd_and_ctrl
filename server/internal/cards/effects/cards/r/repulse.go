package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repulse — Instant {2}{U} (Wilds of Eldraine Commander, #5368):
//
//	"Return target creature to its owner's hand.
//	 Draw a card."
//
// Leave in the Dust's body over a creature target
// (bounceTargetThenDrawACard): the draw is the bounce's continuation, so
// it happens whether or not the creature reached its owner's hand (a
// commander may go to the command zone instead), and a target that
// became illegal in response makes the whole spell do nothing, the draw
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f7ef1d1-c441-41ba-b7d6-6b7727ffcf04",
		Name:         "Repulse",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.BounceTargetThenDrawACard,
	})
}
